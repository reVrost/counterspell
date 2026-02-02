package models

// Session represents a chat session.
type Session struct {
	ID               string  `json:"id"`
	AgentBackend     string  `json:"agent_backend"`
	ExternalID       *string `json:"external_id,omitempty"`
	BackendSessionID *string `json:"backend_session_id,omitempty"`
	Title            *string `json:"title,omitempty"`
	MessageCount     int64   `json:"message_count"`
	LastMessageAt    *int64  `json:"last_message_at,omitempty"`
	CreatedAt        int64   `json:"created_at"`
	UpdatedAt        int64   `json:"updated_at"`
}

// SessionMessage represents a single message/event in a session.
type SessionMessage struct {
	ID         string  `json:"id"`
	SessionID  string  `json:"session_id"`
	Sequence   int64   `json:"sequence"`
	Role       string  `json:"role"`
	Kind       string  `json:"kind"`
	Content    *string `json:"content,omitempty"`
	ToolName   *string `json:"tool_name,omitempty"`
	ToolCallID *string `json:"tool_call_id,omitempty"`
	RawJSON    string  `json:"raw_json"`
	CreatedAt  int64   `json:"created_at"`
}
