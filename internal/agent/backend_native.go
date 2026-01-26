package agent

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/revrost/counterspell/internal/agent/tools"
	"github.com/revrost/counterspell/internal/llm"
)

// ErrProviderRequired is returned when a NativeBackend is created without a provider.
var ErrProviderRequired = errors.New("agent: llm.Provider is required for native backend")

// Compile-time interface check
var _ Backend = (*NativeBackend)(nil)

// NativeBackend wraps the Go-based Runner to implement Backend.
//
// Architecture:
//   - Brain (LLM calls, context management): Runs natively in Go for full
//     goroutine/channel benefits
//   - Hands (file ops, bash): Will be sandboxed via bubblewrap (TODO)
//
// This separation gives us the best of both worlds: low-latency native
// orchestration with secure sandboxed tool execution.
type NativeBackend struct {
	runner *Runner
}

// NativeBackendOption configures a NativeBackend.
type NativeBackendOption func(*nativeBackendConfig)

type nativeBackendConfig struct {
	provider llm.Provider
	workDir  string
	callback StreamCallback
}

// WithProvider sets the LLM provider.
func WithProvider(p llm.Provider) NativeBackendOption {
	return func(c *nativeBackendConfig) {
		c.provider = p
	}
}

// WithWorkDir sets the working directory for file operations.
func WithWorkDir(dir string) NativeBackendOption {
	return func(c *nativeBackendConfig) {
		c.workDir = dir
	}
}

// WithCallback sets the event streaming callback.
func WithCallback(cb StreamCallback) NativeBackendOption {
	return func(c *nativeBackendConfig) {
		c.callback = cb
	}
}

// NewNativeBackend creates a native Go agent backend.
//
// Example:
//
//	backend, err := NewNativeBackend(
//	    WithProvider(llm.NewAnthropicProvider(apiKey)),
//	    WithWorkDir("/path/to/workspace"),
//	    WithCallback(func(e StreamEvent) { ... }),
//	)
func NewNativeBackend(opts ...NativeBackendOption) (*NativeBackend, error) {
	cfg := &nativeBackendConfig{
		workDir: ".",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.provider == nil {
		return nil, ErrProviderRequired
	}

	runner := NewRunner(cfg.provider, cfg.workDir, cfg.callback)
	return &NativeBackend{runner: runner}, nil
}

// --- Backend interface ---

// Run executes a new task.
func (b *NativeBackend) Run(ctx context.Context, task string) error {
	return b.runner.Run(ctx, task)
}

// Send continues the conversation with a follow-up message.
// func (b *NativeBackend) Send(ctx context.Context, message string) error {
// 	return b.runner.Continue(ctx, message)
// }

// Close releases resources (no-op for native, context cancellation handles cleanup).
func (b *NativeBackend) Close() error {
	return nil
}

// --- StatefulBackend interface ---

// GetState returns the conversation history as JSON.
func (b *NativeBackend) GetState() string {
	return b.runner.GetMessageHistory()
}

// RestoreState sets the conversation history from JSON.
func (b *NativeBackend) RestoreState(stateJSON string) error {
	return b.runner.SetMessageHistory(stateJSON)
}

// --- IntrospectableBackend interface ---

// Messages returns the raw conversation history.
func (b *NativeBackend) Messages() []Message {
	// Return a copy to prevent mutation
	history := b.runner.GetMessageHistory()
	if history == "" {
		return nil
	}
	var msgs []Message
	// Safe to ignore error - will return nil on failure
	_ = json.Unmarshal([]byte(history), &msgs)
	return msgs
}

// FinalMessage returns the accumulated assistant response text.
func (b *NativeBackend) FinalMessage() string {
	return b.runner.GetFinalMessage()
}

// Todos returns the current task list.
func (b *NativeBackend) Todos() []tools.TodoItem {
	return b.runner.GetTodoState().GetTodos()
}

// --- Describable interface ---

// Info returns backend metadata.
func (b *NativeBackend) Info() BackendInfo {
	return BackendInfo{
		Type:    BackendNative,
		Version: "1.0.0",
	}
}

// --- Accessors for advanced usage ---

// Runner returns the underlying Runner for direct access.
// Use sparingly - prefer the Backend interface methods.
func (b *NativeBackend) Runner() *Runner {
	return b.runner
}
