// Package agent implements a simple coding agent loop.
//
// AGENT LOOP CORE CONCEPT:
//
// An agent loop enables LLMs to interact with tools through a cycle:
//
//   1. User input → API call (with available tools)
//   2. Response contains either:
//      - text: assistant speaks to user
//      - tool_use: assistant wants to call a function
//   3. If tool_use: execute tool, send result back to API
//   4. Repeat steps 2-3 until assistant returns only text
//   5. Present final text to user
//
// This allows the model to multi-step reason: read files, run commands,
// gather context, then formulate a final answer.
//
// KEY COMPONENTS:
// - Tools: Functions the LLM can call (read, write, edit, glob, grep, bash)
// - Schema: Tool definitions sent to the LLM so it knows what to call
// - Messages: Conversation history including tool calls and results
// - Loop: Keep calling API until no more tool_use blocks
package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	apiURL   = "https://api.anthropic.com/v1/messages"
	model    = "claude-opus-4-5"
	apiVer   = "2023-06-01"
	maxToken = 8192
)

var (
	bold  = "\033[1m"
	dim   = "\033[2m"
	blue  = "\033[34m"
	cyan  = "\033[36m"
	green = "\033[32m"
	red   = "\033[31m"
	reset = "\033[0m"
)

// ============================================================================
// TYPES
// ============================================================================

// ToolFunc is the signature for all tool implementations.
// Takes arguments from the LLM and returns the result as a string.
type ToolFunc func(map[string]any) string

// Tool represents a function the LLM can call.
// - Description: What the tool does (shown to LLM)
// - Schema: Parameter types and whether optional (trailing "?" means optional)
// - Func: The actual Go function to execute
type Tool struct {
	Description string
	Schema      map[string]any
	Func        ToolFunc
}

// Message represents a single message in the conversation.
// The agent appends messages with both text and tool calls.
type Message struct {
	Role    string         `json:"role"` // "user" or "assistant"
	Content []ContentBlock `json:"content"`
}

// ContentBlock can be one of three types:
// - text: Assistant response text
// - tool_use: Assistant requesting to call a tool
// - tool_result: Result sent back after tool execution (from us to API)
type ContentBlock struct {
	// For all blocks
	Type string `json:"type"`

	// For text blocks
	Text string `json:"text,omitempty"`

	// For tool_use blocks (assistant calling a tool)
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`
	ID    string         `json:"id,omitempty"` // Links to tool_result

	// For tool_result blocks (result sent back to assistant)
	ToolUseID string `json:"tool_use_id,omitempty"` // Links to tool_use ID
	Content   string `json:"content,omitempty"`     // Tool output
}

// APIRequest is what we send to Anthropic's API.
type APIRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`    // System prompt (context for the assistant)
	Messages  []Message `json:"messages"`  // Conversation history
	Tools     []ToolDef `json:"tools"`     // Tools available to the assistant
}

// ToolDef is the schema for a single tool, sent to the LLM.
type ToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema defines tool parameters in JSON Schema format.
type InputSchema struct {
	Type       string                 `json:"type"`                 // Always "object"
	Properties map[string]interface{} `json:"properties"`          // Param name -> {type: "..."}
	Required   []string               `json:"required"`             // Names of required params
}

// APIResponse is what we get back from Anthropic's API.
type APIResponse struct {
	Content []ContentBlock `json:"content"` // Assistant's response
}

// ============================================================================
// TOOLS REGISTRY
// ============================================================================

// makeTools returns all available tools.
// Each tool has a description, parameter schema, and implementation.
func makeTools() map[string]Tool {
	return map[string]Tool{
		"read": {
			Description: "Read file with line numbers (file path, not directory)",
			Schema: map[string]any{
				"path":   "string",   // Required
				"offset": "number?",  // Optional
				"limit":  "number?",  // Optional
			},
			Func: toolRead,
		},
		"write": {
			Description: "Write content to file",
			Schema: map[string]any{
				"path":    "string", // Required
				"content": "string", // Required
			},
			Func: toolWrite,
		},
		"edit": {
			Description: "Replace old with new in file (old must be unique unless all=true)",
			Schema: map[string]any{
				"path": "string",    // Required
				"old":  "string",    // Required
				"new":  "string",    // Required
				"all":  "boolean?",  // Optional
			},
			Func: toolEdit,
		},
		"glob": {
			Description: "Find files by pattern, sorted by mtime",
			Schema: map[string]any{
				"pat":  "string",  // Required
				"path": "string?", // Optional
			},
			Func: toolGlob,
		},
		"grep": {
			Description: "Search files for regex pattern",
			Schema: map[string]any{
				"pat":  "string",  // Required
				"path": "string?", // Optional
			},
			Func: toolGrep,
		},
		"bash": {
			Description: "Run shell command",
			Schema: map[string]any{
				"cmd": "string", // Required
			},
			Func: toolBash,
		},
	}
}

// ============================================================================
// TOOL SCHEMA GENERATION
// ============================================================================

// makeSchema converts our Tool definitions to API-compatible ToolDef.
// This maps our simple "string?" notation to JSON Schema format.
func makeSchema(tools map[string]Tool) []ToolDef {
	result := []ToolDef{}
	for name, tool := range tools {
		props := map[string]interface{}{}
		required := []string{}

		// Process each parameter in the tool's schema
		for paramName, paramType := range tool.Schema {
			typeStr, ok := paramType.(string)
			if !ok {
				continue
			}

			// Remove "?" suffix to get base type
			baseType := strings.TrimSuffix(typeStr, "?")

			// Convert our type names to JSON Schema types
			resultType := "string"
			if baseType == "number" {
				resultType = "integer"
			}
			if baseType == "boolean" {
				resultType = "boolean"
			}

			props[paramName] = map[string]any{"type": resultType}

			// If no "?" suffix, parameter is required
			if !strings.HasSuffix(typeStr, "?") {
				required = append(required, paramName)
			}
		}

		result = append(result, ToolDef{
			Name:        name,
			Description: tool.Description,
			InputSchema: InputSchema{
				Type:       "object",
				Properties: props,
				Required:   required,
			},
		})
	}
	return result
}

// ============================================================================
// TOOL EXECUTION
// ============================================================================

// runTool executes a tool by name with given arguments.
// Includes panic recovery so tool crashes don't crash the agent.
func runTool(name string, args map[string]any, tools map[string]Tool) string {
	tool, ok := tools[name]
	if !ok {
		return fmt.Sprintf("error: unknown tool %s", name)
	}
	// Recover from panics in tool implementations
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("error in %s: %v\n", name, r)
		}
	}()
	return tool.Func(args)
}

// toolRead reads a file and formats it with line numbers.
// Supports offset and limit for reading file chunks.
func toolRead(args map[string]any) string {
	path := args["path"].(string)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	lines := strings.Split(string(data), "\n")

	// Parse optional parameters
	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}
	limit := len(lines)
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Calculate slice bounds
	end := offset + limit
	if end > len(lines) {
		end = len(lines)
	}

	var sb strings.Builder
	for i := offset; i < end; i++ {
		sb.WriteString(fmt.Sprintf("%4d| %s\n", i+1, lines[i]))
	}
	return sb.String()
}

// toolWrite writes content to a file.
// Overwrites existing file or creates new one.
func toolWrite(args map[string]any) string {
	path := args["path"].(string)
	content := args["content"].(string)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return "ok"
}

// toolEdit replaces old string with new string in a file.
// Safety check: old string must be unique unless all=true is passed.
func toolEdit(args map[string]any) string {
	path := args["path"].(string)
	oldStr := args["old"].(string)
	newStr := args["new"].(string)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	text := string(data)

	// Safety check: verify old string exists
	if !strings.Contains(text, oldStr) {
		return "error: old_string not found"
	}

	// Check for uniqueness unless all=true
	count := strings.Count(text, oldStr)
	doAll := false
	if a, ok := args["all"].(bool); ok {
		doAll = a
	}
	if !doAll && count > 1 {
		return fmt.Sprintf("error: old_string appears %d times, must be unique (use all=true)", count)
	}

	// Perform replacement
	var replacement string
	if doAll {
		replacement = strings.ReplaceAll(text, oldStr, newStr)
	} else {
		replacement = strings.Replace(text, oldStr, newStr, 1)
	}

	if err := os.WriteFile(path, []byte(replacement), 0644); err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return "ok"
}

// toolGlob finds files matching a pattern, sorted by modification time.
// Most recently modified files appear first.
func toolGlob(args map[string]any) string {
	pat := args["pat"].(string)
	basePath := "."
	if p, ok := args["path"].(string); ok {
		basePath = p
	}

	fullPat := filepath.Join(basePath, pat)
	matches, err := filepath.Glob(fullPat)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	// Collect file info for sorting by modification time
	type fileInfo struct {
		path string
		mtime time.Time
	}
	fileInfos := []fileInfo{}
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{match, info.ModTime()})
	}

	// Sort: most recently modified first
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].mtime.After(fileInfos[j].mtime)
	})

	if len(fileInfos) == 0 {
		return "none"
	}

	var sb strings.Builder
	for _, fi := range fileInfos {
		sb.WriteString(fi.path + "\n")
	}
	return sb.String()
}

// toolGrep searches files for regex pattern matches.
// Returns results in format: filepath:line:content
func toolGrep(args map[string]any) string {
	pat := args["pat"].(string)
	basePath := "."
	if p, ok := args["path"].(string); ok {
		basePath = p
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		return fmt.Sprintf("error: invalid regex: %v", err)
	}

	hits := []string{}
	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		// Skip errors and directories
		if err != nil || info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(data), "\n")
		for lineNum, line := range lines {
			if re.MatchString(line) {
				hits = append(hits, fmt.Sprintf("%s:%d:%s", path, lineNum+1, strings.TrimSpace(line)))
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	if len(hits) == 0 {
		return "none"
	}

	// Limit results to prevent overwhelming output
	maxHits := 50
	if len(hits) > maxHits {
		hits = hits[:maxHits]
	}
	return strings.Join(hits, "\n")
}

// toolBash executes a shell command and returns stdout + stderr.
func toolBash(args map[string]any) string {
	cmdStr := args["cmd"].(string)
	cmd := exec.Command("bash", "-c", cmdStr)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String() + stderr.String()

	// Include exit error in output if command failed
	if err != nil {
		output += fmt.Sprintf("\n(error: %v)", err)
	}

	if strings.TrimSpace(output) == "" {
		return "(empty)"
	}
	return output
}

// ============================================================================
// API COMMUNICATION
// ============================================================================

// callAPI sends a request to Anthropic's API and returns the response.
// Includes tool definitions so the assistant knows what it can call.
func callAPI(messages []Message, systemPrompt string, tools map[string]Tool) (*APIResponse, error) {
	req := APIRequest{
		Model:     model,
		MaxTokens: maxToken,
		System:    systemPrompt,
		Messages:  messages,
		Tools:     makeSchema(tools), // Convert tools to API schema
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := createHTTPRequest(apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &apiResp, nil
}

// createHTTPRequest builds the HTTP request with proper headers.
// API key is read from ANTHROPIC_API_KEY environment variable.
func createHTTPRequest(url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
	req.Header.Set("anthropic-version", apiVer)
	return req, nil
}

// ============================================================================
// UI HELPERS
// ============================================================================

// separator returns a visual separator line that adapts to terminal width.
func separator() string {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w < width {
		width = w
	}
	return dim + strings.Repeat("─", width) + reset
}

// renderMarkdown converts basic markdown to ANSI colors.
// Currently only handles **bold** text.
func renderMarkdown(text string) string {
	re := regexp.MustCompile(`\*\*(.+?)\*\*`)
	return re.ReplaceAllString(text, bold+"$1"+reset)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// ============================================================================
// AGENT LOOP (CORE LOGIC)
// ============================================================================

// Run starts the interactive agent loop.
//
// FLOW:
//   1. Read user input from stdin
//   2. Append user message to conversation history
//   3. Call API with conversation history
//   4. Process response:
//      - If text block: display to user
//      - If tool_use block: execute tool, send result back to API (go to step 3)
//   5. If no tool_use blocks: loop complete, wait for next user input
//
// This is the "agent loop" - the core pattern that enables autonomous tool use.
func Run() {
	cwd, _ := os.Getwd()
	fmt.Printf("%snanocode%s | %s%s | %s%s%s\n\n", bold, reset, dim, model, dim, cwd, reset)

	tools := makeTools()          // Available tools for the LLM
	messages := []Message{}        // Conversation history
	systemPrompt := fmt.Sprintf("Concise coding assistant. cwd: %s", cwd)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println(separator())
		fmt.Printf("%s%s%s❯%s ", bold, blue, reset, reset)
		if !scanner.Scan() {
			break
		}
		userInput := strings.TrimSpace(scanner.Text())
		fmt.Println(separator())

		// Handle empty input
		if userInput == "" {
			continue
		}

		// Handle exit commands
		if userInput == "/q" || userInput == "exit" {
			break
		}

		// Handle clear conversation command
		if userInput == "/c" {
			messages = []Message{}
			fmt.Printf("%s⏺ Cleared conversation%s\n", green, reset)
			continue
		}

		// Append user message to conversation history
		messages = append(messages, Message{
			Role: "user",
			Content: []ContentBlock{
				{Type: "text", Text: userInput},
			},
		})

		// ============================================================
		// AGENTIC LOOP: Keep calling API until no more tool calls
		// ============================================================
		for {
			apiResp, err := callAPI(messages, systemPrompt, tools)
			if err != nil {
				fmt.Printf("%s⏺ Error: %v%s\n", red, err, reset)
				break
			}

			contentBlocks := apiResp.Content
			toolResults := []ContentBlock{}

			// Process each block in the assistant's response
			for _, block := range contentBlocks {
				// Text block: assistant speaking to user
				if block.Type == "text" {
					fmt.Printf("\n%s⏺%s %s\n", cyan, reset, renderMarkdown(block.Text))
				}

				// Tool use block: assistant wants to execute a function
				if block.Type == "tool_use" {
					toolName := block.Name
					toolArgs := block.Input

					// Show what tool is being called (preview first arg)
					argPreview := ""
					if len(toolArgs) > 0 {
						for _, v := range toolArgs {
							argPreview = fmt.Sprintf("%v", v)
							break
						}
					}
					if len(argPreview) > 50 {
						argPreview = argPreview[:50]
					}
					fmt.Printf("\n%s⏺ %s%s(%s%s%s)\n", green, capitalize(toolName), reset, dim, argPreview, reset)

					// Execute the tool
					result := runTool(toolName, toolArgs, tools)

					// Show result preview (first line, truncated)
					resultLines := strings.Split(result, "\n")
					preview := resultLines[0]
					if len(preview) > 60 {
						preview = preview[:60] + "..."
					}
					if len(resultLines) > 1 {
						preview += fmt.Sprintf(" ... +%d lines", len(resultLines)-1)
					}
					fmt.Printf("  %s⎿  %s%s\n", dim, preview, reset)

					// Store result to send back to assistant
					// This links back to the tool_use block via ID
					toolResults = append(toolResults, ContentBlock{
						Type:        "tool_result",
						ToolUseID:   block.ID,
						Content:     result,
					})
				}
			}

			// Append assistant's response to conversation history
			// This includes both text and tool_use blocks
			messages = append(messages, Message{Role: "assistant", Content: contentBlocks})

			// If no tools were called, we're done with this turn
			// The assistant has finished and is waiting for user input
			if len(toolResults) == 0 {
				break
			}

			// Tool results exist: send them back to the API
			// The assistant will use this information to continue its reasoning
			messages = append(messages, Message{Role: "user", Content: toolResults})
		}
		fmt.Println()
	}
}
