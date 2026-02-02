package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"sync"

	"github.com/revrost/counterspell/internal/agent/tools"
)

// Compile-time interface check
var _ Backend = (*CodexBackend)(nil)

// ErrNoCodexBinaryPath is returned when the codex binary cannot be found.
var ErrNoCodexBinaryPath = errors.New("agent: codex binary not found in PATH")

// ErrCodexUnsupported is returned when codex execution isn't implemented yet.
var ErrCodexUnsupported = errors.New("agent: codex backend not implemented")

// CodexBackend is a placeholder backend for the Codex CLI.
// It currently supports event parsing for tests and plumbing, but does not run the CLI.
type CodexBackend struct {
	binaryPath string
	workDir    string
	callback   StreamCallback
	apiKey     string
	baseURL    string
	model      string
	sessionID  string

	mu           sync.Mutex
	cmd          *exec.Cmd
	cancel       context.CancelFunc
	finalMessage string
	messages     []Message
}

// CodexOption configures a CodexBackend.
type CodexOption func(*CodexBackend)

// WithCodexBinaryPath sets the path to the codex binary.
// Defaults to "codex" (found via PATH).
func WithCodexBinaryPath(path string) CodexOption {
	return func(b *CodexBackend) {
		b.binaryPath = path
	}
}

// WithCodexWorkDir sets the working directory for the codex process.
func WithCodexWorkDir(dir string) CodexOption {
	return func(b *CodexBackend) {
		b.workDir = dir
	}
}

// WithCodexCallback sets the event streaming callback.
func WithCodexCallback(cb StreamCallback) CodexOption {
	return func(b *CodexBackend) {
		b.callback = cb
	}
}

// WithCodexAPIKey sets the API key for Codex.
func WithCodexAPIKey(key string) CodexOption {
	return func(b *CodexBackend) {
		b.apiKey = key
	}
}

// WithCodexBaseURL sets a custom API base URL.
func WithCodexBaseURL(url string) CodexOption {
	return func(b *CodexBackend) {
		b.baseURL = url
	}
}

// WithCodexModel sets the model to use.
func WithCodexModel(model string) CodexOption {
	return func(b *CodexBackend) {
		b.model = model
	}
}

// WithCodexSessionID sets the Codex session ID to continue.
func WithCodexSessionID(sessionID string) CodexOption {
	return func(b *CodexBackend) {
		b.sessionID = sessionID
	}
}

// NewCodexBackend creates a Codex CLI backend.
func NewCodexBackend(opts ...CodexOption) (*CodexBackend, error) {
	b := &CodexBackend{
		binaryPath: "codex",
		workDir:    ".",
	}
	for _, opt := range opts {
		opt(b)
	}

	if _, err := exec.LookPath(b.binaryPath); err != nil {
		return nil, ErrNoCodexBinaryPath
	}

	return b, nil
}

// Run executes a new task via Codex CLI.
func (b *CodexBackend) Run(ctx context.Context, task string) error {
	return ErrCodexUnsupported
}

// Close terminates any running codex process.
func (b *CodexBackend) Close() error {
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

// GetState returns the conversation history as JSON.
func (b *CodexBackend) GetState() string {
	return ""
}

// RestoreState initializes the backend with previously saved state.
func (b *CodexBackend) RestoreState(stateJSON string) error {
	return nil
}

// Messages returns the raw conversation history.
func (b *CodexBackend) Messages() []Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]Message(nil), b.messages...)
}

// FinalMessage returns the accumulated assistant response text.
func (b *CodexBackend) FinalMessage() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.finalMessage
}

// Todos returns the current task list.
func (b *CodexBackend) Todos() []tools.TodoItem {
	return nil
}

// Info returns backend type and capabilities.
func (b *CodexBackend) Info() BackendInfo {
	return BackendInfo{Type: BackendCodex}
}

type codexEvent struct {
	Type     string     `json:"type"`
	ThreadID string     `json:"thread_id"`
	Item     *codexItem `json:"item"`
}

type codexItem struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Text    string `json:"text"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
}

func (b *CodexBackend) parseOutput(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		switch event.Type {
		case "thread.started":
			if event.ThreadID != "" {
				b.emit(StreamEvent{Type: "session", SessionID: event.ThreadID})
			}
		case "item.started":
			if event.Item != nil && event.Item.Type == "command_execution" {
				b.emit(StreamEvent{Type: EventTool, Content: event.Item.Command})
			}
		case "item.completed":
			if event.Item == nil {
				continue
			}
			switch event.Item.Type {
			case "command_execution":
				content := event.Item.Stdout
				if content == "" {
					content = event.Item.Stderr
				}
				b.emit(StreamEvent{Type: EventToolResult, Content: content})
			case "agent_message":
				b.emit(StreamEvent{Type: EventText, Content: event.Item.Text})
			}
		case "turn.completed":
			b.emit(StreamEvent{Type: EventDone})
		}
	}
}

func (b *CodexBackend) emit(event StreamEvent) {
	if b.callback != nil {
		b.callback(event)
	}

	if event.Type == EventText && event.Content != "" {
		b.mu.Lock()
		b.finalMessage += event.Content
		b.messages = append(b.messages, Message{
			Role:    "assistant",
			Content: []ContentBlock{{Type: "text", Text: event.Content}},
		})
		b.mu.Unlock()
	}
}
