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

	"github.com/lithammer/shortuuid/v4"
	"github.com/revrost/counterspell/internal/agent/tools"
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
	binaryPath    string
	workDir       string
	apiKey        string
	baseURL       string
	model         string
	sessionID     string // Claude Code session ID
	streamCtx     context.Context
	events        chan<- StreamEvent
	streamMsgID   string
	streamMsgRole string
	streamText    string

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

// WithSessionID sets the Claude Code session ID to continue.
func WithSessionID(sessionID string) ClaudeCodeOption {
	return func(b *ClaudeCodeBackend) {
		b.sessionID = sessionID
	}
}

// NewClaudeCodeBackend creates a Claude Code CLI backend.
//
// Example:
//
//	backend, err := NewClaudeCodeBackend(
//	    WithBinaryPath("/usr/local/bin/claude"),
//	    WithClaudeWorkDir("/path/to/workspace"),
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
	stream := b.Stream(ctx, task)
	return drainStream(ctx, stream)
}

// Stream executes a task via Claude Code CLI and returns a stream of events.
func (b *ClaudeCodeBackend) Stream(ctx context.Context, task string) *Stream {
	events := make(chan StreamEvent, 32)
	done := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(done)
		b.setStream(ctx, events)
		defer b.clearStream()

		done <- b.execute(ctx, task)
	}()

	return &Stream{Events: events, Done: done}
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

func (b *ClaudeCodeBackend) setStream(ctx context.Context, events chan<- StreamEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.streamCtx = ctx
	b.events = events
}

func (b *ClaudeCodeBackend) clearStream() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.streamCtx = nil
	b.events = nil
}

// --- Internal execution ---

func (b *ClaudeCodeBackend) execute(ctx context.Context, prompt string) error {
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

	cmd, err := b.buildCmd(ctx, prompt)
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
				b.emit(StreamEvent{Type: EventError, Error: line})
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
			b.appendStreamText(line)
			continue
		}

		b.processClaudeEvent(event)
	}
}

func (b *ClaudeCodeBackend) processClaudeEvent(event map[string]any) {
	eventType, _ := event["type"].(string)

	switch eventType {
	case "system":
		// Capture session ID from system event
		if sessionID, ok := event["session_id"].(string); ok && sessionID != "" {
			changed := false
			b.mu.Lock()
			if b.sessionID != sessionID {
				b.sessionID = sessionID
				changed = true
			}
			b.mu.Unlock()
			if changed {
				slog.Info("[CLAUDE-CODE] Session ID detected", "session_id", sessionID)
				// Emit session ID for orchestrator to save
				b.emit(StreamEvent{Type: EventSession, SessionID: sessionID})
			}
		}

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
				}
			}
		}

	case "assistant":
		// Flush any streaming text before handling a full assistant message
		streamText := b.flushStreamText()
		if streamText != "" {
			b.mu.Lock()
			b.messages = append(b.messages, Message{Role: "assistant", Content: []ContentBlock{{Type: "text", Text: streamText}}})
			b.mu.Unlock()
			b.emit(StreamEvent{
				Type:      EventContentEnd,
				MessageID: b.streamMsgID,
				BlockType: "text",
				Block:     &ContentBlock{Type: "text", Text: streamText},
			})
			b.endStreamMessage()
		}

		if message, ok := event["message"].(map[string]any); ok {
			if content, ok := message["content"].([]any); ok {
				msgID := shortuuid.New()
				b.emit(StreamEvent{Type: EventMessageStart, MessageID: msgID, Role: "assistant"})
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
								b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "text", Block: &ContentBlock{Type: "text"}})
								b.emit(StreamEvent{Type: EventContentEnd, MessageID: msgID, BlockType: "text", Block: &ContentBlock{Type: "text", Text: text}})
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
							b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "tool_use", Block: &ContentBlock{Type: "tool_use", Name: name, ID: id, Input: input}})
							b.emit(StreamEvent{Type: EventContentEnd, MessageID: msgID, BlockType: "tool_use", Block: &ContentBlock{Type: "tool_use", Name: name, ID: id, Input: input}})
						}
					}
				}
				if len(blocks) > 0 {
					b.mu.Lock()
					b.messages = append(b.messages, Message{Role: "assistant", Content: blocks})
					b.mu.Unlock()
				}
				b.emit(StreamEvent{Type: EventMessageEnd, MessageID: msgID, Role: "assistant"})
			}
		}

	case "content_block_delta":
		if delta, ok := event["delta"].(map[string]any); ok {
			if text, ok := delta["text"].(string); ok {
				b.appendStreamText(text)
			}
		}

	case "tool_use":
		streamText := b.flushStreamText()
		if streamText != "" {
			b.mu.Lock()
			b.messages = append(b.messages, Message{Role: "assistant", Content: []ContentBlock{{Type: "text", Text: streamText}}})
			b.mu.Unlock()
			b.emit(StreamEvent{
				Type:      EventContentEnd,
				MessageID: b.streamMsgID,
				BlockType: "text",
				Block:     &ContentBlock{Type: "text", Text: streamText},
			})
			b.endStreamMessage()
		}
		name, _ := event["name"].(string)
		id, _ := event["id"].(string)
		input, _ := event["input"].(map[string]any)

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
		msgID := shortuuid.New()
		b.emit(StreamEvent{Type: EventMessageStart, MessageID: msgID, Role: "assistant"})
		b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "tool_use", Block: &ContentBlock{Type: "tool_use", Name: name, ID: id, Input: input}})
		b.emit(StreamEvent{Type: EventContentEnd, MessageID: msgID, BlockType: "tool_use", Block: &ContentBlock{Type: "tool_use", Name: name, ID: id, Input: input}})
		b.emit(StreamEvent{Type: EventMessageEnd, MessageID: msgID, Role: "assistant"})

	case "tool_result":
		streamText := b.flushStreamText()
		if streamText != "" {
			b.mu.Lock()
			b.messages = append(b.messages, Message{Role: "assistant", Content: []ContentBlock{{Type: "text", Text: streamText}}})
			b.mu.Unlock()
			b.emit(StreamEvent{
				Type:      EventContentEnd,
				MessageID: b.streamMsgID,
				BlockType: "text",
				Block:     &ContentBlock{Type: "text", Text: streamText},
			})
			b.endStreamMessage()
		}
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
		msgID := shortuuid.New()
		b.emit(StreamEvent{Type: EventMessageStart, MessageID: msgID, Role: "user"})
		b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "tool_result", Block: &ContentBlock{Type: "tool_result", ToolUseID: toolUseID, Content: content}})
		b.emit(StreamEvent{Type: EventContentEnd, MessageID: msgID, BlockType: "tool_result", Block: &ContentBlock{Type: "tool_result", ToolUseID: toolUseID, Content: content}})
		b.emit(StreamEvent{Type: EventMessageEnd, MessageID: msgID, Role: "user"})

	case "result":
		// Check if this is an error result
		isError, _ := event["is_error"].(bool)
		resultText, _ := event["result"].(string)
		if isError {
			slog.Error("[CLAUDE-CODE] Result error", "result", resultText)
			b.emit(StreamEvent{Type: EventError, Error: resultText})
		} else {
			streamText := b.flushStreamText()
			if streamText != "" {
				b.mu.Lock()
				b.messages = append(b.messages, Message{Role: "assistant", Content: []ContentBlock{{Type: "text", Text: streamText}}})
				b.mu.Unlock()
				b.emit(StreamEvent{
					Type:      EventContentEnd,
					MessageID: b.streamMsgID,
					BlockType: "text",
					Block:     &ContentBlock{Type: "text", Text: streamText},
				})
				b.endStreamMessage()
			}
			b.emit(StreamEvent{Type: EventDone})
		}
	}
}

func (b *ClaudeCodeBackend) emit(event StreamEvent) {
	b.mu.Lock()
	ctx := b.streamCtx
	events := b.events
	b.mu.Unlock()
	if events == nil {
		return
	}
	if ctx == nil {
		events <- event
		return
	}
	select {
	case <-ctx.Done():
		return
	case events <- event:
	}
}

func (b *ClaudeCodeBackend) startStreamMessage(role string) string {
	b.mu.Lock()
	if b.streamMsgID != "" {
		id := b.streamMsgID
		b.mu.Unlock()
		return id
	}
	id := shortuuid.New()
	b.streamMsgID = id
	b.streamMsgRole = role
	b.mu.Unlock()
	b.emit(StreamEvent{Type: EventMessageStart, MessageID: id, Role: role})
	return id
}

func (b *ClaudeCodeBackend) endStreamMessage() {
	b.mu.Lock()
	id := b.streamMsgID
	role := b.streamMsgRole
	b.streamMsgID = ""
	b.streamMsgRole = ""
	b.mu.Unlock()
	if id != "" {
		b.emit(StreamEvent{Type: EventMessageEnd, MessageID: id, Role: role})
	}
}

func (b *ClaudeCodeBackend) appendStreamText(delta string) {
	if delta == "" {
		return
	}
	msgID := b.startStreamMessage("assistant")
	start := false
	b.mu.Lock()
	if b.streamText == "" {
		start = true
	}
	b.streamText += delta
	b.finalMessage += delta
	b.mu.Unlock()
	if start {
		b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "text", Block: &ContentBlock{Type: "text"}})
	}
	b.emit(StreamEvent{Type: EventContentDelta, MessageID: msgID, BlockType: "text", Delta: delta})
}

func (b *ClaudeCodeBackend) flushStreamText() string {
	b.mu.Lock()
	text := b.streamText
	b.streamText = ""
	b.mu.Unlock()
	return text
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
	return nil
}

// SessionID returns the current Claude Code session ID.
func (b *ClaudeCodeBackend) SessionID() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sessionID
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
func (b *ClaudeCodeBackend) buildCmd(ctx context.Context, prompt string) (*exec.Cmd, error) {
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

	// Use existing session ID if provided, otherwise use --continue flag
	if b.sessionID != "" {
		args = append(args, "-r")
		args = append(args, b.sessionID)
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
