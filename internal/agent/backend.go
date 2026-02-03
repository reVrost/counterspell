// Package agent provides agent backend abstractions for executing LLM-powered tasks.
//
// Architecture:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                         Backend Interface                        │
//	│  Core: Run, Send, Close                                         │
//	│  State: GetState, RestoreState                                  │
//	│  Introspection: Messages, FinalMessage, Todos                   │
//	└─────────────────────────────────────────────────────────────────┘
//	           │                                    │
//	           ▼                                    ▼
//	┌─────────────────────┐              ┌─────────────────────┐
//	│   NativeBackend     │              │  ClaudeCodeBackend  │
//	│   (Go agent loop)   │              │  (CLI wrapper)      │
//	└─────────────────────┘              └─────────────────────┘
//
// Design principles:
//   - Single interface: All backends must implement everything for UI consistency
//   - UI-coupled: Backend features map directly to UI panels (chat, todos, diff)
//   - Context-aware: All blocking operations accept context
package agent

import (
	"context"

	"github.com/revrost/counterspell/internal/agent/tools"
)

// Backend is the interface for agent task execution.
// All methods are required - the UI depends on full implementation.
//
// Lifecycle: New() -> RestoreState? -> Stream/SendStream -> (repeat) -> Close
//
// Thread safety: Implementations must be safe for concurrent method calls,
// but callers should not call Stream/SendStream concurrently on the same backend.
type Backend interface {
	// --- Core execution ---

	// Stream executes a task and returns a stream of events.
	// The caller must drain Events until Done is received.
	Stream(ctx context.Context, task string) *Stream

	// Run executes a task and blocks until completion.
	// It drains the stream internally; prefer Stream for responsiveness.
	Run(ctx context.Context, task string) error

	// // Send continues the conversation with a follow-up message.
	// // Must be called after a successful Run. Returns error if no session active.
	// Send(ctx context.Context, message string) error

	// Close releases resources. Safe to call multiple times.
	Close() error

	// --- State management (required for conversation persistence) ---

	// GetState returns the conversation history as JSON.
	// Returns empty string if no conversation active.
	GetState() string

	// RestoreState initializes the backend with previously saved state.
	// Must be called before Run if restoring a session.
	RestoreState(stateJSON string) error

	// --- Introspection (required for UI rendering) ---

	// Messages returns the raw conversation history.
	Messages() []Message

	// FinalMessage returns the accumulated assistant response text.
	FinalMessage() string

	// Todos returns the current task list.
	Todos() []tools.TodoItem

	// --- Metadata ---

	// Info returns backend type and capabilities.
	Info() BackendInfo
}

// --- Backend info for introspection ---

// BackendType identifies the agent implementation.
type BackendType string

const (
	BackendNative     BackendType = "native"      // Go-based agent loop
	BackendClaudeCode BackendType = "claude-code" // Claude Code CLI
	BackendCodex      BackendType = "codex"       // OpenAI Codex CLI
)

// BackendInfo describes a backend implementation.
type BackendInfo struct {
	Type    BackendType
	Version string
}
