package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sync"
)

// Compile-time interface checks
var (
	_ Backend     = (*ClaudeCodeBackend)(nil)
	_ Describable = (*ClaudeCodeBackend)(nil)
)

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

	mu             sync.Mutex
	cmd            *exec.Cmd
	cancel         context.CancelFunc
	messageHistory string
	finalMessage   string
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

// Send continues the conversation with a follow-up message.
func (b *ClaudeCodeBackend) Send(ctx context.Context, message string) error {
	return b.execute(ctx, message, true)
}

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

	// Build command args for JSON streaming mode
	args := []string{
		"--print",
		"--output-format", "stream-json",
	}

	if isContinuation {
		args = append(args, "--continue")
	}

	args = append(args, "--", prompt)

	// TODO: Wrap with bubblewrap for sandboxing
	// args = b.wrapWithBubblewrap(args)

	cmd := exec.CommandContext(ctx, b.binaryPath, args...)
	cmd.Dir = b.workDir

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

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start claude: %w", err)
	}

	// Parse output in background
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		b.parseOutput(bufio.NewScanner(stdout))
	}()

	go func() {
		defer wg.Done()
		b.parseStderr(bufio.NewScanner(stderr))
	}()

	err = cmd.Wait()
	wg.Wait()

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

func (b *ClaudeCodeBackend) parseStderr(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			b.emit(StreamEvent{Type: EventError, Content: line})
		}
	}
}

func (b *ClaudeCodeBackend) processClaudeEvent(event map[string]any) {
	eventType, _ := event["type"].(string)

	switch eventType {
	case "assistant":
		if message, ok := event["message"].(map[string]any); ok {
			if content, ok := message["content"].([]any); ok {
				for _, block := range content {
					if blockMap, ok := block.(map[string]any); ok {
						if text, ok := blockMap["text"].(string); ok {
							b.mu.Lock()
							b.finalMessage += text
							b.mu.Unlock()
							b.emit(StreamEvent{Type: EventText, Content: text})
						}
					}
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
		input, _ := event["input"].(map[string]any)
		inputJSON, _ := json.Marshal(input)
		b.emit(StreamEvent{
			Type:    EventTool,
			Tool:    name,
			Args:    string(inputJSON),
			Content: fmt.Sprintf("Running %s", name),
		})

	case "tool_result":
		content, _ := event["content"].(string)
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		b.emit(StreamEvent{Type: EventResult, Content: content})

	case "result":
		b.emit(StreamEvent{Type: EventDone, Content: "Task completed"})
	}
}

func (b *ClaudeCodeBackend) emit(event StreamEvent) {
	if b.callback != nil {
		b.callback(event)
	}
}

// --- Describable interface ---

// Info returns backend metadata.
func (b *ClaudeCodeBackend) Info() BackendInfo {
	return BackendInfo{
		Type:         BackendClaudeCode,
		Version:      "1.0.0",
		Capabilities: []string{}, // No optional interfaces fully supported
	}
}

// --- Limited state access ---
// These methods exist for API compatibility but have limitations.

// GetFinalMessage returns the accumulated assistant response text.
func (b *ClaudeCodeBackend) GetFinalMessage() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.finalMessage
}
