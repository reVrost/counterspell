package models

import "time"

// TaskStatus represents the status of a task in the workflow.
type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusReview     TaskStatus = "review"
	StatusDone       TaskStatus = "done"
)

// Task represents a work item in the system.
type Task struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"project_id"`
	Title          string     `json:"title"`
	Intent         string     `json:"intent"`
	Status         TaskStatus `json:"status"`
	Position       int        `json:"position"`
	AgentOutput    string     `json:"agent_output,omitempty"`
	GitDiff        string     `json:"git_diff,omitempty"`
	MessageHistory string     `json:"message_history,omitempty"` // JSON serialized agent message history
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// LogLevel represents the severity level of an agent log entry.
type LogLevel string

const (
	LogInfo    LogLevel = "info"
	LogPlan    LogLevel = "plan"
	LogCode    LogLevel = "code"
	LogError   LogLevel = "error"
	LogSuccess LogLevel = "success"
)

// AgentLog represents a log entry from agent execution.
type AgentLog struct {
	ID        int64     `json:"id"`
	TaskID    string    `json:"task_id"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// EventType represents the type of server-sent event.
type EventType string

const (
	EventTypeLog            EventType = "log"
	EventTypeStatus         EventType = "status"
	EventTypeStatusChange   EventType = "status_change"
	EventTypeError          EventType = "error"
	EventTypeAgentUpdate    EventType = "agent_update"
	EventTypeTodo           EventType = "todo"
	EventTypeTaskCreated    EventType = "task_created"
	EventTypeProjectCreated EventType = "project_created"
	EventTypeProjectUpdated EventType = "project_updated"
	EventTypeProjectDeleted EventType = "project_deleted"
)

// Event represents a server-sent event for real-time updates.
type Event struct {
	ID          int64     `json:"id"`           // Sequence number for deduplication
	TaskID      string    `json:"task_id"`
	Type        EventType `json:"type"`
	HTMLPayload string    `json:"html_payload"`
}
