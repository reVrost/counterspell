package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/revrost/code/counterspell/internal/agent/tools"
)

// Compile-time interface check
var _ Backend = (*ClaudeCodeBackend)(nil)

// ErrNoBinaryPath is returned when the claude binary cannot be found.
var ErrNoBinaryPath = errors.New("agent: claude binary not found in PATH")

// ClaudeCodeBackend wraps the Claude Code CLI as a Backend.
//
// Architecture:
//   - The entire claude binary is intended to run inside bubblewrap
//   - No native Go processing - we're just a subprocess manager
//   - Events are parsed from the CLI's JSON streaming output
//
// Limitations compared to NativeBackend:
//   - No goroutine benefits (it's a subprocess)
//   - State management is limited (CLI manages its own state)
//   - IntrospectableBackend not fully supported
type ClaudeCodeBackend struct {
	binaryPath string
	workDir    string
	callback   StreamCallback
	apiKey     string
	baseURL    string
	model      string

	mu           sync.Mutex
	cmd          *exec.Cmd
	cancel       context.CancelFunc
	finalMessage string
	messages     []Message // Track conversation for UI updates
}

// ClaudeCodeOption configures a ClaudeCodeBackend.
type ClaudeCodeOption func(*ClaudeCodeBackend)

// WithBinaryPath sets the path to the claude binary.
// Defaults to "claude" (found via PATH).
func WithBinaryPath(path string) ClaudeCodeOption {
	return func(b *ClaudeCodeBackend) {
		b.binaryPath = path
	}
}

// WithClaudeWorkDir sets the working directory for the claude process.
func WithClaudeWorkDir(dir string) ClaudeCodeOption {
	return func(b *ClaudeCodeBackend) {
		b.workDir = dir
	}
}

// WithClaudeCallback sets the event streaming callback.
func WithClaudeCallback(cb StreamCallback) ClaudeCodeOption {
	return func(b *ClaudeCodeBackend) {
		b.callback = cb
	}
}

// WithAPIKey sets the Anthropic API key for Claude Code CLI.
func WithAPIKey(key string) ClaudeCodeOption {
	return func(b *ClaudeCodeBackend) {
		b.apiKey = key
	}
}

// WithBaseURL sets a custom API base URL (e.g., for Z.AI compatibility).
func WithBaseURL(url string) ClaudeCodeOption {
	return func(b *ClaudeCodeBackend) {
		b.baseURL = url
	}
}

// WithModel sets the model to use.
func WithModel(model string) ClaudeCodeOption {
	return func(b *ClaudeCodeBackend) {
		b.model = model
	}
}

// NewClaudeCodeBackend creates a Claude Code CLI backend.
//
// Example:
//
//	backend, err := NewClaudeCodeBackend(
//	    WithBinaryPath("/usr/local/bin/claude"),
//	    WithClaudeWorkDir("/path/to/workspace"),
//	    WithClaudeCallback(func(e StreamEvent) { ... }),
//	)
func NewClaudeCodeBackend(opts ...ClaudeCodeOption) (*ClaudeCodeBackend, error) {
	b := &ClaudeCodeBackend{
		binaryPath: "claude",
		workDir:    ".",
	}
	for _, opt := range opts {
		opt(b)
	}

	// Verify binary exists
	if _, err := exec.LookPath(b.binaryPath); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNoBinaryPath, b.binaryPath)
	}

	return b, nil
}

// --- Backend interface ---

// Run executes a new task via Claude Code CLI.
func (b *ClaudeCodeBackend) Run(ctx context.Context, task string) error {
	return b.execute(ctx, task, false)
}

// // Send continues the conversation with a follow-up message.
// func (b *ClaudeCodeBackend) Send(ctx context.Context, message string) error {
// 	return b.execute(ctx, message, true)
// }

// Close terminates any running claude process.
func (b *ClaudeCodeBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
	}
	if b.cmd != nil && b.cmd.Process != nil {
		return b.cmd.Process.Kill()
	}
	return nil
}

// --- Internal execution ---

func (b *ClaudeCodeBackend) execute(ctx context.Context, prompt string, isContinuation bool) error {
	ctx, cancel := context.WithCancel(ctx)

	b.mu.Lock()
	b.cancel = cancel
	b.mu.Unlock()

	// Add user message to history and emit immediately
	b.mu.Lock()
	b.messages = append(b.messages, Message{
		Role:    "user",
		Content: []ContentBlock{{Type: "text", Text: prompt}},
	})
	b.mu.Unlock()
	b.emitMessages()

	cmd, err := b.buildCmd(ctx, prompt, isContinuation)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.cmd = cmd
	b.mu.Unlock()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	slog.Info("[CLAUDE-CODE] Starting command", "binary", b.binaryPath, "args", cmd.Args, "workdir", b.workDir, "api_key_len", len(b.apiKey), "base_url", b.baseURL)

	if err := cmd.Start(); err != nil {
		slog.Error("[CLAUDE-CODE] Failed to start command", "binary", b.binaryPath, "args", cmd.Args, "workdir", b.workDir, "error", err)
		return fmt.Errorf("start claude: %w", err)
	}

	// Parse output in background
	var wg sync.WaitGroup
	var stderrLines []string
	var stderrMu sync.Mutex

	wg.Go(func() {
		b.parseOutput(bufio.NewScanner(stdout))
	})

	wg.Go(func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				slog.Warn("[CLAUDE-CODE] stderr", "line", line)
				stderrMu.Lock()
				stderrLines = append(stderrLines, line)
				stderrMu.Unlock()
				b.emit(StreamEvent{Type: EventError, Content: line})
			}
		}
	})

	err = cmd.Wait()
	wg.Wait()

	if err != nil {
		stderrMu.Lock()
		stderrContent := strings.Join(stderrLines, "\n")
		stderrMu.Unlock()
		if stderrContent != "" {
			return fmt.Errorf("%w: %s", err, stderrContent)
		}
	}
	return err
}

func (b *ClaudeCodeBackend) parseOutput(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Not JSON, emit as text
			b.emit(StreamEvent{Type: EventText, Content: line})
			continue
		}

		b.processClaudeEvent(event)
	}
}

func (b *ClaudeCodeBackend) processClaudeEvent(event map[string]any) {
	eventType, _ := event["type"].(string)

	switch eventType {
	case "user":
		// User message from the CLI
		if message, ok := event["message"].(map[string]any); ok {
			if content, ok := message["content"].([]any); ok {
				var blocks []ContentBlock
				for _, block := range content {
					if blockMap, ok := block.(map[string]any); ok {
						if text, ok := blockMap["text"].(string); ok {
							blocks = append(blocks, ContentBlock{Type: "text", Text: text})
						}
					}
				}
				if len(blocks) > 0 {
					b.mu.Lock()
					b.messages = append(b.messages, Message{Role: "user", Content: blocks})
					b.mu.Unlock()
					b.emitMessages()
				}
			}
		}

	case "assistant":
		if message, ok := event["message"].(map[string]any); ok {
			if content, ok := message["content"].([]any); ok {
				var blocks []ContentBlock
				for _, block := range content {
					if blockMap, ok := block.(map[string]any); ok {
						blockType, _ := blockMap["type"].(string)
						switch blockType {
						case "text":
							if text, ok := blockMap["text"].(string); ok {
								b.mu.Lock()
								b.finalMessage += text
								b.mu.Unlock()
								blocks = append(blocks, ContentBlock{Type: "text", Text: text})
								b.emit(StreamEvent{Type: EventText, Content: text})
							}
						case "tool_use":
							name, _ := blockMap["name"].(string)
							id, _ := blockMap["id"].(string)
							input, _ := blockMap["input"].(map[string]any)
							blocks = append(blocks, ContentBlock{
								Type:  "tool_use",
								Name:  name,
								ID:    id,
								Input: input,
							})
							inputJSON, _ := json.Marshal(input)
							b.emit(StreamEvent{
								Type:    EventTool,
								Tool:    name,
								Args:    string(inputJSON),
								Content: fmt.Sprintf("Running %s", name),
							})
						}
					}
				}
				if len(blocks) > 0 {
					b.mu.Lock()
					b.messages = append(b.messages, Message{Role: "assistant", Content: blocks})
					b.mu.Unlock()
					b.emitMessages()
				}
			}
		}

	case "content_block_delta":
		if delta, ok := event["delta"].(map[string]any); ok {
			if text, ok := delta["text"].(string); ok {
				b.mu.Lock()
				b.finalMessage += text
				b.mu.Unlock()
				b.emit(StreamEvent{Type: EventText, Content: text})
			}
		}

	case "tool_use":
		name, _ := event["name"].(string)
		id, _ := event["id"].(string)
		input, _ := event["input"].(map[string]any)
		inputJSON, _ := json.Marshal(input)

		// Add tool use to messages
		b.mu.Lock()
		b.messages = append(b.messages, Message{
			Role: "assistant",
			Content: []ContentBlock{{
				Type:  "tool_use",
				Name:  name,
				ID:    id,
				Input: input,
			}},
		})
		b.mu.Unlock()
		b.emitMessages()

		b.emit(StreamEvent{
			Type:    EventTool,
			Tool:    name,
			Args:    string(inputJSON),
			Content: fmt.Sprintf("Running %s", name),
		})

	case "tool_result":
		content, _ := event["content"].(string)
		toolUseID, _ := event["tool_use_id"].(string)

		// Add tool result to messages
		b.mu.Lock()
		b.messages = append(b.messages, Message{
			Role: "user",
			Content: []ContentBlock{{
				Type:      "tool_result",
				ToolUseID: toolUseID,
				Content:   content,
			}},
		})
		b.mu.Unlock()
		b.emitMessages()

		if len(content) > 200 {
			content = content[:200] + "..."
		}
		b.emit(StreamEvent{Type: EventToolResult, Content: content})

	case "result":
		// Check if this is an error result
		isError, _ := event["is_error"].(bool)
		resultText, _ := event["result"].(string)
		if isError {
			slog.Error("[CLAUDE-CODE] Result error", "result", resultText)
			b.emit(StreamEvent{Type: EventError, Content: resultText})
		} else {
			b.emit(StreamEvent{Type: EventDone, Content: "Task completed"})
		}
	}
}

func (b *ClaudeCodeBackend) emit(event StreamEvent) {
	slog.Debug("[CLAUDE-CODE] emit", "type", event.Type, "content_len", len(event.Content), "has_callback", b.callback != nil)
	if b.callback != nil {
		b.callback(event)
	}
}

func (b *ClaudeCodeBackend) emitMessages() {
	if b.callback == nil {
		slog.Debug("[CLAUDE-CODE] emitMessages skipped - no callback")
		return
	}
	b.mu.Lock()
	msgs := make([]Message, len(b.messages))
	copy(msgs, b.messages)
	b.mu.Unlock()

	data, err := json.Marshal(msgs)
	if err != nil {
		slog.Error("[CLAUDE-CODE] emitMessages failed to marshal", "error", err)
		return
	}
	slog.Debug("[CLAUDE-CODE] emitMessages", "msg_count", len(msgs), "data_len", len(data))
}

// --- Describable interface ---

// Info returns backend metadata.
func (b *ClaudeCodeBackend) Info() BackendInfo {
	return BackendInfo{
		Type:    BackendClaudeCode,
		Version: "1.0.0",
	}
}

// --- StatefulBackend interface ---

// GetState returns the conversation history as JSON.
func (b *ClaudeCodeBackend) GetState() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	data, err := json.Marshal(b.messages)
	if err != nil {
		return ""
	}
	return string(data)
}

// RestoreState initializes the backend with previously saved state.
func (b *ClaudeCodeBackend) RestoreState(stateJSON string) error {
	if stateJSON == "" {
		return nil
	}
	var msgs []Message
	if err := json.Unmarshal([]byte(stateJSON), &msgs); err != nil {
		return err
	}
	b.mu.Lock()
	b.messages = msgs
	b.mu.Unlock()
	// Emit restored messages to update UI
	b.emitMessages()
	return nil
}

// --- IntrospectableBackend interface ---

// Messages returns the raw conversation history.
func (b *ClaudeCodeBackend) Messages() []Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Return a copy to prevent mutation
	result := make([]Message, len(b.messages))
	copy(result, b.messages)
	return result
}

// FinalMessage returns the accumulated assistant response text.
func (b *ClaudeCodeBackend) FinalMessage() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.finalMessage
}

// Todos returns the current task list (empty for Claude Code backend).
func (b *ClaudeCodeBackend) Todos() []tools.TodoItem {
	// Claude Code manages its own todos internally, we don't track them
	return nil
}

// buildCmd constructs the exec.Cmd for running Claude Code.
func (b *ClaudeCodeBackend) buildCmd(ctx context.Context, prompt string, isContinuation bool) (*exec.Cmd, error) {
	// Build command args for JSON streaming mode
	// Note: --verbose is required for stream-json output format
	args := []string{
		"--print",
		"--verbose",
		"--output-format", "stream-json",
		"--dangerously-skip-permissions",
	}

	if b.model != "" {
		args = append(args, "--model", b.model)
	}

	if isContinuation {
		args = append(args, "--continue")
	}

	args = append(args, "--", prompt)

	// TODO: Wrap with bubblewrap for sandboxing
	// args = b.wrapWithBubblewrap(args)

	cmd := exec.CommandContext(ctx, b.binaryPath, args...)
	cmd.Dir = b.workDir

	// Set environment variables for API authentication
	// For Z.AI/OpenRouter, use ANTHROPIC_AUTH_TOKEN and ANTHROPIC_BASE_URL
	env := os.Environ()
	if b.baseURL != "" {
		env = append(env, "ANTHROPIC_BASE_URL="+b.baseURL)
		// Custom providers use ANTHROPIC_AUTH_TOKEN
		if b.apiKey != "" {
			env = append(env, "ANTHROPIC_AUTH_TOKEN="+b.apiKey)
		}
		// Important: Explicitly blank ANTHROPIC_API_KEY to prevent conflicts (required by OpenRouter)
		env = append(env, "ANTHROPIC_API_KEY=")
	} else if b.apiKey != "" {
		// Standard Anthropic API
		env = append(env, "ANTHROPIC_API_KEY="+b.apiKey)
	}

	// Set model environment variables for GLM-4.7 compatibility
	if b.model == "glm-4.7" {
		env = append(env, "ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-4.7")
		env = append(env, "ANTHROPIC_DEFAULT_SONNET_MODEL=glm-4.7")
		env = append(env, "ANTHROPIC_DEFAULT_OPUS_MODEL=glm-4.7")
	}

	cmd.Env = env
	return cmd, nil
}
