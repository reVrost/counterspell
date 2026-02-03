package agent

import "github.com/revrost/counterspell/internal/agent/tools"

// Stream represents an asynchronous stream of agent events.
type Stream struct {
	Events <-chan StreamEvent
	Done   <-chan error
}

// StreamEventType identifies the type of streaming event.
type StreamEventType string

const (
	EventMessageStart StreamEventType = "message_start"
	EventContentStart StreamEventType = "content_start"
	EventContentDelta StreamEventType = "content_delta"
	EventContentEnd   StreamEventType = "content_end"
	EventMessageEnd   StreamEventType = "message_end"
	EventTodo         StreamEventType = "todo"
	EventError        StreamEventType = "error"
	EventDone         StreamEventType = "done"
	EventSession      StreamEventType = "session"
)

// StreamEvent represents a single event in the agent execution.
type StreamEvent struct {
	Type      StreamEventType  `json:"type"`
	MessageID string           `json:"message_id,omitempty"`
	Role      string           `json:"role,omitempty"`
	BlockType string           `json:"block_type,omitempty"`
	Delta     string           `json:"delta,omitempty"`
	Block     *ContentBlock    `json:"block,omitempty"`
	SessionID string           `json:"session_id,omitempty"`
	Todos     []tools.TodoItem `json:"todos,omitempty"`
	Error     string           `json:"error,omitempty"`
}
