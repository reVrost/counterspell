package models

import "time"

// Event represents a system event for SSE broadcasting.
type Event struct {
	ID      int64     `json:"id"`
	TaskID  string    `json:"task_id"`
	UserID  string    `json:"user_id"` // "default" for local-first mode
	Type    string    `json:"type"`   // "task_update", "agent_update", "system"
	Data    string    `json:"data"`   // JSON payload
	CreatedAt time.Time `json:"created_at"`
}
