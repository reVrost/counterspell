package models

import "time"

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusPlanning   TaskStatus = "planning"
	StatusInProgress TaskStatus = "in_progress"
	StatusReview     TaskStatus = "review"
	StatusDone       TaskStatus = "done"
	StatusFailed     TaskStatus = "failed"
)

// Task represents a work item.
type Task struct {
	ID              string    `json:"id"`
	ProjectID       *string   `json:"project_id,omitempty"`
	MachineID       string    `json:"machine_id"`
	Title           string    `json:"title"`
	Intent          string    `json:"intent"`
	Status          string    `json:"status"`
	Position        int64     `json:"position"`
	CurrentStep     string    `json:"current_step,omitempty"`
	AssignedAgentID string    `json:"assigned_agent_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Machine represents a worker instance.
type Machine struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Mode         string    `json:"mode"`                   // "local" or "cloud"
	Capabilities string    `json:"capabilities,omitempty"` // JSON string
	LastSeenAt   time.Time `json:"last_seen_at"`
}

// Agent represents an agent configuration.
type Agent struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	SystemPrompt string    `json:"system_prompt"`
	Tools        string    `json:"tools"` // JSON string
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AgentRun represents an execution of an agent.
type AgentRun struct {
	ID             string    `json:"id"`
	TaskID         string    `json:"task_id"`
	Step           string    `json:"step"`
	AgentID        string    `json:"agent_id,omitempty"`
	Status         string    `json:"status"` // pending, running, completed, failed
	Input          string    `json:"input,omitempty"`
	Output         string    `json:"output,omitempty"`
	MessageHistory string    `json:"message_history,omitempty"` // JSON string
	ArtifactPath   string    `json:"artifact_path,omitempty"`
	Error          string    `json:"error,omitempty"`
	StartedAt      time.Time `json:"started_at,omitempty"`
	CompletedAt    time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// Settings represents application settings.
type Settings struct {
	OpenRouterKey string    `json:"openrouter_key,omitempty"`
	ZaiKey        string    `json:"zai_key,omitempty"`
	AnthropicKey  string    `json:"anthropic_key,omitempty"`
	OpenAIKey     string    `json:"openai_key,omitempty"`
	AgentBackend  string    `json:"agent_backend"` // "native", "openai", "anthropic", "openrouter", "zai"
	UpdatedAt     time.Time `json:"updated_at"`
}
