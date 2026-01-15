// Package agent implements a simple coding agent loop.
package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/revrost/code/counterspell/internal/llm"
)

// Event types for streaming
const (
	EventPlan     = "plan"
	EventTool     = "tool"
	EventResult   = "result"
	EventText     = "text"
	EventError    = "error"
	EventDone     = "done"
	EventMessages = "messages" // Full message history update
)

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

// StreamEvent represents a single event in the agent execution.
type StreamEvent struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	Tool     string `json:"tool,omitempty"`
	Args     string `json:"args,omitempty"`
	Messages string `json:"messages,omitempty"` // JSON message history for EventMessages
}

// StreamCallback is called for each event during agent execution.
type StreamCallback func(event StreamEvent)

// Runner executes agent tasks with streaming output.
type Runner struct {
	provider       llm.Provider
	llmCaller      LLMCaller
	workDir        string
	callback       StreamCallback
	systemPrompt   string
	finalMessage   string
	messageHistory []Message
}

// NewRunner creates a new agent runner.
func NewRunner(provider llm.Provider, workDir string, callback StreamCallback) *Runner {
	return &Runner{
		provider:     provider,
		llmCaller:    NewLLMCaller(provider),
		workDir:      workDir,
		callback:     callback,
		systemPrompt: fmt.Sprintf("You are a coding assistant. Work directory: %s. Be concise. Make changes directly.", workDir),
	}
}

// GetFinalMessage returns the accumulated final message from the agent.
func (r *Runner) GetFinalMessage() string {
	return r.finalMessage
}

// GetMessageHistory returns the message history as JSON string.
func (r *Runner) GetMessageHistory() string {
	data, err := json.Marshal(r.messageHistory)
	if err != nil {
		return ""
	}
	return string(data)
}

// SetMessageHistory sets the initial message history from JSON string.
func (r *Runner) SetMessageHistory(historyJSON string) error {
	if historyJSON == "" {
		return nil
	}
	return json.Unmarshal([]byte(historyJSON), &r.messageHistory)
}

// Run executes the agent loop for a given task.
func (r *Runner) Run(ctx context.Context, task string) error {
	return r.runWithMessage(ctx, task, false)
}

// Continue resumes the agent loop with a new follow-up message.
func (r *Runner) Continue(ctx context.Context, followUpMessage string) error {
	return r.runWithMessage(ctx, followUpMessage, true)
}

// runWithMessage is the core loop that handles both new runs and continuations.
func (r *Runner) runWithMessage(ctx context.Context, userMessage string, isContinuation bool) error {
	tools := r.makeTools()

	// Use existing message history or start fresh
	messages := r.messageHistory
	if messages == nil {
		messages = []Message{}
	}

	if isContinuation {
		r.emit(StreamEvent{Type: EventPlan, Content: "Continuing with: " + userMessage})
	} else {
		r.emit(StreamEvent{Type: EventPlan, Content: "Analyzing task: " + userMessage})
	}

	// Add user message
	messages = append(messages, Message{
		Role: "user",
		Content: []ContentBlock{
			{Type: "text", Text: userMessage},
		},
	})

	// Emit immediately so user message appears in UI right away
	r.emitMessages(messages)

	// Agent loop
	for {
		select {
		case <-ctx.Done():
			r.messageHistory = messages
			return ctx.Err()
		default:
		}

		r.emit(StreamEvent{Type: EventPlan, Content: "Calling LLM API..."})
		resp, err := r.llmCaller.Call(messages, tools, r.systemPrompt)
		if err != nil {
			r.messageHistory = messages
			r.emit(StreamEvent{Type: EventError, Content: err.Error()})
			return err
		}
		r.emit(StreamEvent{Type: EventPlan, Content: fmt.Sprintf("Received response with %d content blocks", len(resp.Content))})

		// Log the raw response for debugging
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Printf("\n=== LLM API RESPONSE ===\n%s\n=== END RESPONSE ===\n\n", string(respJSON))
		for i, block := range resp.Content {
			fmt.Printf("Block %d: type=%s, has_text=%v, has_name=%v, id=%s\n", i, block.Type, block.Text != "", block.Name != "", block.ID)
		}

		toolResults := []ContentBlock{}

		// Immediately add assistant message and emit so UI shows response right away
		messages = append(messages, Message{Role: "assistant", Content: resp.Content})
		r.emitMessages(messages)

		for _, block := range resp.Content {
			if block.Type == "text" && block.Text != "" {
				r.emit(StreamEvent{Type: EventText, Content: block.Text})
				r.finalMessage += block.Text
			}

			if block.Type == "tool_use" {
				argsJSON, _ := json.Marshal(block.Input)
				r.emit(StreamEvent{
					Type:    EventTool,
					Tool:    block.Name,
					Args:    string(argsJSON),
					Content: fmt.Sprintf("Running %s", block.Name),
				})

				result := r.runTool(block.Name, block.Input, tools)

				// Truncate result for display
				displayResult := result
				if len(displayResult) > 200 {
					displayResult = displayResult[:200] + "..."
				}
				r.emit(StreamEvent{
					Type:    EventResult,
					Tool:    block.Name,
					Content: displayResult,
				})

				toolResults = append(toolResults, ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   result,
				})
			}
		}

		if len(toolResults) == 0 {
			r.emit(StreamEvent{Type: EventPlan, Content: "No more tools to run, completing task"})
			break
		}

		r.emit(StreamEvent{Type: EventPlan, Content: fmt.Sprintf("Running %d tool result(s) through agent loop", len(toolResults))})
		messages = append(messages, Message{Role: "user", Content: toolResults})

		// Emit with tool results so UI shows them
		r.emitMessages(messages)
	}

	// Store message history for future continuations
	r.messageHistory = messages

	r.emit(StreamEvent{Type: EventDone, Content: "Task completed"})
	return nil
}

func (r *Runner) emit(event StreamEvent) {
	if r.callback != nil {
		r.callback(event)
	}
}

func (r *Runner) emitMessages(messages []Message) {
	data, err := json.Marshal(messages)
	if err != nil {
		return
	}
	r.emit(StreamEvent{
		Type:     EventMessages,
		Messages: string(data),
	})
}

func (r *Runner) runTool(name string, args map[string]any, tools map[string]Tool) string {
	tool, ok := tools[name]
	if !ok {
		return fmt.Sprintf("error: unknown tool %s", name)
	}
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Printf("error in %s: %v\n", name, rec)
		}
	}()
	return tool.Func(args)
}

// makeTools creates tools that work in the runner's work directory.
func (r *Runner) makeTools() map[string]Tool {
	return map[string]Tool{
		"read": {
			Description: "Read file with line numbers",
			Schema: map[string]any{
				"path":   "string",
				"offset": "number?",
				"limit":  "number?",
			},
			Func: r.toolRead,
		},
		"write": {
			Description: "Write content to file",
			Schema: map[string]any{
				"path":    "string",
				"content": "string",
			},
			Func: r.toolWrite,
		},
		"edit": {
			Description: "Replace old with new in file",
			Schema: map[string]any{
				"path": "string",
				"old":  "string",
				"new":  "string",
				"all":  "boolean?",
			},
			Func: r.toolEdit,
		},
		"glob": {
			Description: "Find files by pattern",
			Schema: map[string]any{
				"pat":  "string",
				"path": "string?",
			},
			Func: r.toolGlob,
		},
		"grep": {
			Description: "Search files for regex pattern",
			Schema: map[string]any{
				"pat":  "string",
				"path": "string?",
			},
			Func: r.toolGrep,
		},
		"bash": {
			Description: "Run shell command",
			Schema: map[string]any{
				"cmd": "string",
			},
			Func: r.toolBash,
		},
		"ls": {
			Description: "List directory contents",
			Schema: map[string]any{
				"path": "string?",
			},
			Func: r.toolLs,
		},
	}
}

func (r *Runner) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(r.workDir, path)
}

func (r *Runner) toolRead(args map[string]any) string {
	path := r.resolvePath(args["path"].(string))
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	lines := strings.Split(string(data), "\n")

	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}
	limit := len(lines)
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	end := min(offset+limit, len(lines))

	var sb strings.Builder
	for i := offset; i < end; i++ {
		fmt.Fprintf(&sb, "%4d| %s\n", i+1, lines[i])
	}
	return sb.String()
}

func (r *Runner) toolWrite(args map[string]any) string {
	path := r.resolvePath(args["path"].(string))
	content := args["content"].(string)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return "ok"
}

func (r *Runner) toolEdit(args map[string]any) string {
	path := r.resolvePath(args["path"].(string))
	oldStr := args["old"].(string)
	newStr := args["new"].(string)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	text := string(data)

	if !strings.Contains(text, oldStr) {
		return "error: old_string not found"
	}

	count := strings.Count(text, oldStr)
	doAll := false
	if a, ok := args["all"].(bool); ok {
		doAll = a
	}
	if !doAll && count > 1 {
		return fmt.Sprintf("error: old_string appears %d times, use all=true", count)
	}

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

func (r *Runner) toolGlob(args map[string]any) string {
	pat := args["pat"].(string)
	basePath := r.workDir
	if p, ok := args["path"].(string); ok {
		basePath = r.resolvePath(p)
	}

	fullPat := filepath.Join(basePath, pat)
	matches, err := filepath.Glob(fullPat)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	type fileInfo struct {
		path  string
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

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].mtime.After(fileInfos[j].mtime)
	})

	if len(fileInfos) == 0 {
		return "none"
	}

	var sb strings.Builder
	for _, fi := range fileInfos {
		// Make path relative to workdir for cleaner output
		rel, _ := filepath.Rel(r.workDir, fi.path)
		if rel == "" {
			rel = fi.path
		}
		sb.WriteString(rel + "\n")
	}
	return sb.String()
}

func (r *Runner) toolGrep(args map[string]any) string {
	pat := args["pat"].(string)
	basePath := r.workDir
	if p, ok := args["path"].(string); ok {
		basePath = r.resolvePath(p)
	}

	re, err := regexp.Compile(pat)
	if err != nil {
		return fmt.Sprintf("error: invalid regex: %v", err)
	}

	hits := []string{}
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Skip binary files and hidden directories
		if strings.Contains(path, "/.git/") || strings.Contains(path, "/node_modules/") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(data), "\n")
		for lineNum, line := range lines {
			if re.MatchString(line) {
				rel, _ := filepath.Rel(r.workDir, path)
				if rel == "" {
					rel = path
				}
				hits = append(hits, fmt.Sprintf("%s:%d:%s", rel, lineNum+1, strings.TrimSpace(line)))
			}
		}
		return nil
	})

	if len(hits) == 0 {
		return "none"
	}

	if len(hits) > 50 {
		hits = hits[:50]
	}
	return strings.Join(hits, "\n")
}

func (r *Runner) toolBash(args map[string]any) string {
	cmdStr := args["cmd"].(string)
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = r.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String() + stderr.String()

	if err != nil {
		output += fmt.Sprintf("\n(exit: %v)", err)
	}

	if strings.TrimSpace(output) == "" {
		return "(empty)"
	}
	return output
}

func (r *Runner) toolLs(args map[string]any) string {
	path := r.workDir
	if p, ok := args["path"].(string); ok {
		path = r.resolvePath(p)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	var sb strings.Builder
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		sb.WriteString(name + "\n")
	}
	return sb.String()
}
