// Package agent provides agent backend abstractions for executing LLM-powered tasks.
//
// Architecture:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                         Backend Interface                        │
//	│  (Run, Send, Close)                                             │
//	└─────────────────────────────────────────────────────────────────┘
//	           │                                    │
//	           ▼                                    ▼
//	┌─────────────────────┐              ┌─────────────────────┐
//	│   NativeBackend     │              │  ClaudeCodeBackend  │
//	│   (Go agent loop)   │              │  (CLI wrapper)      │
//	│                     │              │                     │
//	│  Brain: Native Go   │              │  Fully sandboxed    │
//	│  Hands: Sandboxed   │              │  (bwrap entire CLI) │
//	└─────────────────────┘              └─────────────────────┘
//
// Design principles:
//   - Small interface: Only essential methods for task execution
//   - Capability-based: Optional interfaces for extended features (use type assertions)
//   - Backward compatible: Existing StreamEvent/StreamCallback still work
//   - Context-aware: All blocking operations accept context
package agent

import (
	"context"

	"github.com/revrost/code/counterspell/internal/agent/tools"
)

// Backend is the core interface for agent task execution.
// Implementations handle the specifics of how tasks are run.
//
// Lifecycle: New() -> Run/Send -> (repeat) -> Close
//
// Thread safety: Implementations must be safe for concurrent method calls,
// but callers should not call Run/Send concurrently on the same backend.
type Backend interface {
	// Run executes a task and streams events via the configured callback.
	// Blocks until completion; cancel via context.
	Run(ctx context.Context, task string) error

	// Send continues the conversation with a follow-up message.
	// Must be called after a successful Run. Returns error if no session active.
	Send(ctx context.Context, message string) error

	// Close releases resources. Safe to call multiple times.
	// Implementations should handle graceful shutdown.
	Close() error
}

// --- Optional capability interfaces ---
// Backends may implement these for extended functionality.
// Use type assertion to check: if sb, ok := backend.(StatefulBackend); ok { ... }

// StatefulBackend can persist and restore conversation state.
// Useful for resuming sessions across process restarts.
type StatefulBackend interface {
	Backend

	// State returns the conversation history as JSON.
	// Returns empty string if no conversation active.
	GetState() string

	// RestoreState initializes the backend with previously saved state.
	// Must be called before Run if restoring a session.
	RestoreState(stateJSON string) error
}

// IntrospectableBackend provides visibility into internal state.
// Primarily for debugging and UI rendering.
type IntrospectableBackend interface {
	Backend

	// Messages returns the raw conversation history.
	Messages() []Message

	// FinalMessage returns the accumulated assistant response text.
	FinalMessage() string

	// Todos returns the current task list (if agent tracks todos).
	Todos() []tools.TodoItem
}

// --- Backend info for introspection ---

// BackendType identifies the agent implementation.
type BackendType string

const (
	BackendNative     BackendType = "native"      // Go-based agent loop
	BackendClaudeCode BackendType = "claude-code" // Claude Code CLI
)

// BackendInfo describes a backend implementation.
type BackendInfo struct {
	Type         BackendType
	Version      string
	Capabilities []string // e.g., "stateful", "introspectable", "streaming"
}

// Describable backends can report their capabilities.
type Describable interface {
	Info() BackendInfo
}
