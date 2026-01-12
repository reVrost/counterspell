package agent

import (
	"bytes"
	"context"
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
)

// Event types for streaming
const (
	EventPlan   = "plan"
	EventTool   = "tool"
	EventResult = "result"
	EventText   = "text"
	EventError  = "error"
	EventDone   = "done"
)

// StreamEvent represents a single event in the agent execution.
type StreamEvent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Tool    string `json:"tool,omitempty"`
	Args    string `json:"args,omitempty"`
}

// StreamCallback is called for each event during agent execution.
type StreamCallback func(event StreamEvent)

// Runner executes agent tasks with streaming output.
type Runner struct {
	apiKey       string
	workDir      string
	callback     StreamCallback
	systemPrompt string
	finalMessage string
}

// NewRunner creates a new agent runner.
func NewRunner(apiKey, workDir string, callback StreamCallback) *Runner {
	return &Runner{
		apiKey:       apiKey,
		workDir:      workDir,
		callback:     callback,
		systemPrompt: fmt.Sprintf("You are a coding assistant. Work directory: %s. Be concise. Make changes directly.", workDir),
	}
}

// GetFinalMessage returns the accumulated final message from the agent.
func (r *Runner) GetFinalMessage() string {
	return r.finalMessage
}

// Run executes the agent loop for a given task.
func (r *Runner) Run(ctx context.Context, task string) error {
	tools := r.makeTools()
	messages := []Message{}

	r.emit(StreamEvent{Type: EventPlan, Content: "Analyzing task: " + task})

	// Add user message
	messages = append(messages, Message{
		Role: "user",
		Content: []ContentBlock{
			{Type: "text", Text: task},
		},
	})

	// Agent loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := r.callAPI(messages, tools)
		if err != nil {
			r.emit(StreamEvent{Type: EventError, Content: err.Error()})
			return err
		}

		toolResults := []ContentBlock{}

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

		messages = append(messages, Message{Role: "assistant", Content: resp.Content})

		if len(toolResults) == 0 {
			break
		}

		messages = append(messages, Message{Role: "user", Content: toolResults})
	}

	r.emit(StreamEvent{Type: EventDone, Content: "Task completed"})
	return nil
}

func (r *Runner) emit(event StreamEvent) {
	if r.callback != nil {
		r.callback(event)
	}
}

func (r *Runner) callAPI(messages []Message, tools map[string]Tool) (*APIResponse, error) {
	req := APIRequest{
		Model:     model,
		MaxTokens: maxToken,
		System:    r.systemPrompt,
		Messages:  messages,
		Tools:     makeSchema(tools),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", r.apiKey)
	httpReq.Header.Set("anthropic-version", apiVer)

	client := &http.Client{Timeout: 120 * time.Second}
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
