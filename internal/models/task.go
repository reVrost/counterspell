package models

import "time"

// TaskStatus represents the status of a task in the workflow.
// Flow: pending → in_progress → review → done
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"     // Task created, not yet started
	StatusInProgress TaskStatus = "in_progress" // Agents are working on the task
	StatusReview     TaskStatus = "review"      // Ready for human review
	StatusDone       TaskStatus = "done"        // Complete
	StatusFailed     TaskStatus = "failed"      // Task failed
)

// Task represents a work item in the system.
type Task struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	Intent      string     `json:"intent"`
	Status      TaskStatus `json:"status"`
	Position    int        `json:"position"`
	CurrentStep string     `json:"current_step,omitempty"` // Current workflow step
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Assignment fields
	AssignedAgentID string `json:"assigned_agent_id,omitempty"`
	AssignedUserID  string `json:"assigned_user_id,omitempty"`
}

// TaskWithProject represents a task with its associated project info.
type TaskWithProject struct {
	Task    *Task
	Project *Project
}

// AgentRunStatus represents the status of an agent run.
type AgentRunStatus string

const (
	RunStatusPending   AgentRunStatus = "pending"
	RunStatusRunning   AgentRunStatus = "running"
	RunStatusCompleted AgentRunStatus = "completed"
	RunStatusFailed    AgentRunStatus = "failed"
)

// AgentRun represents one execution of an agent within a task.
type AgentRun struct {
	ID             string         `json:"id"`
	TaskID         string         `json:"task_id"`
	Step           string         `json:"step"`
	AgentID        string         `json:"agent_id,omitempty"`
	Status         AgentRunStatus `json:"status"`
	Input          string         `json:"input,omitempty"`
	Output         string         `json:"output,omitempty"`
	MessageHistory []Message      `json:"message_history,omitempty"`
	ArtifactPath   string         `json:"artifact_path,omitempty"`
	Error          string         `json:"error,omitempty"`
	StartedAt      *time.Time     `json:"started_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

// Message represents a single message in agent conversation.
type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp,omitempty"`
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
	EventTypeRunUpdate      EventType = "run_update"
)

// Event represents a server-sent event for real-time updates.
type Event struct {
	ID     int64     `json:"id"`
	TaskID string    `json:"task_id"`
	Type   EventType `json:"type"`
	Data   string    `json:"data"`
}
