package models

import "time"

// TaskStatus represents the status of a task in the workflow.
type TaskStatus string

const (
	StatusTodo        TaskStatus = "todo"
	StatusInProgress  TaskStatus = "in_progress"
	StatusReview      TaskStatus = "review"
	StatusHumanReview TaskStatus = "human_review"
	StatusDone        TaskStatus = "done"
)

// Task represents a work item in the system.
type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Intent    string     `json:"intent"`
	Status    TaskStatus `json:"status"`
	Position  int        `json:"position"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
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

// Event represents a server-sent event for real-time updates.
type Event struct {
	TaskID      string `json:"task_id"`
	Type        string `json:"type"` // log, status, error
	HTMLPayload string `json:"html_payload"`
}
