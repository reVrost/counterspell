package models

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
	ID                   string  `json:"id"`
	RepositoryID         *string `json:"repository_id,omitempty"`
	RepositoryName       *string `json:"repository_name,omitempty"`
	SessionID            *string `json:"session_id,omitempty"`
	Title                string  `json:"title"`
	Intent               string  `json:"intent"`
	PromotedSnapshot     *string `json:"promoted_snapshot,omitempty"`
	Status               string  `json:"status"`
	Position             *int64  `json:"position,omitempty"`
	LastAssistantMessage *string `json:"last_assistant_message,omitempty"`
	CreatedAt            int64   `json:"created_at"`
	UpdatedAt            int64   `json:"updated_at"`
}

// AgentRun represents an execution of an agent.
type AgentRun struct {
	ID               string  `json:"id"`
	TaskID           string  `json:"task_id"`
	Prompt           string  `json:"prompt"`
	AgentBackend     string  `json:"agent_backend"`
	Provider         *string `json:"provider,omitempty"`
	Model            *string `json:"model,omitempty"`
	SummaryMessageID *string `json:"summary_message_id,omitempty"`
	BackendSessionID *string `json:"backend_session_id,omitempty"`
	Cost             float64 `json:"cost"`
	MessageCount     int64   `json:"message_count"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	CompletedAt      *int64  `json:"completed_at,omitempty"`
	CreatedAt        int64   `json:"created_at"`
	UpdatedAt        int64   `json:"updated_at"`
}

// Settings represents application settings.
type Settings struct {
	ID            int64   `json:"id"`
	OpenRouterKey *string `json:"openrouter_key,omitempty"`
	ZaiKey        *string `json:"zai_key,omitempty"`
	AnthropicKey  *string `json:"anthropic_key,omitempty"`
	OpenAIKey     *string `json:"openai_key,omitempty"`
	AgentBackend  string  `json:"agent_backend"`
	UpdatedAt     int64   `json:"updated_at"`
}

// Artifact represents a file uploaded by an agent.
type Artifact struct {
	ID        string `json:"id"`
	RunID     string `json:"run_id"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	Version   int64  `json:"version"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// Message represents a chat message for agent conversation history.
type Message struct {
	ID         string  `json:"id"`
	TaskID     string  `json:"task_id"`
	RunID      *string `json:"run_id,omitempty"`
	Role       string  `json:"role"`
	Parts      string  `json:"parts"`
	Model      *string `json:"model,omitempty"`
	Provider   *string `json:"provider,omitempty"`
	Content    string  `json:"content"`
	ToolID     *string `json:"tool_id,omitempty"`
	CreatedAt  int64   `json:"created_at"`
	UpdatedAt  int64   `json:"updated_at"`
	FinishedAt *int64  `json:"finished_at,omitempty"`
}

// AgentRunWithDetails represents an agent run with nested messages and artifacts.
type AgentRunWithDetails struct {
	ID               string     `json:"id"`
	TaskID           string     `json:"task_id"`
	Prompt           string     `json:"prompt"`
	AgentBackend     string     `json:"agent_backend"`
	SummaryMessageID *string    `json:"summary_message_id,omitempty"`
	Cost             float64    `json:"cost"`
	MessageCount     int64      `json:"message_count"`
	PromptTokens     int64      `json:"prompt_tokens"`
	CompletionTokens int64      `json:"completion_tokens"`
	CompletedAt      *int64     `json:"completed_at,omitempty"`
	CreatedAt        int64      `json:"created_at"`
	UpdatedAt        int64      `json:"updated_at"`
	Messages         []Message  `json:"messages,omitempty"`
	Artifacts        []Artifact `json:"artifacts,omitempty"`
}

// TaskResponse represents a detailed task response with all related data.
// Used by the API handler to provide complete task information including messages, git diff, and artifacts.
type TaskResponse struct {
	// Task information
	Task      Task       `json:"task"`
	Messages  []Message  `json:"messages"`
	Artifacts []Artifact `json:"artifacts"`

	// All agent runs with nested messages and artifacts
	AgentRuns []AgentRunWithDetails `json:"agent_runs,omitempty"`

	// Git diff from the worktree (if available)
	GitDiff string `json:"git_diff,omitempty"`
}

// Repository represents a GitHub repository.
type Repository struct {
	ID           string  `json:"id"`
	ConnectionID string  `json:"connection_id"`
	Name         string  `json:"name"`
	FullName     string  `json:"full_name"`
	Owner        string  `json:"owner"`
	IsPrivate    bool    `json:"is_private"`
	HTMLUrl      string  `json:"html_url"`
	CloneUrl     string  `json:"clone_url"`
	LocalPath    *string `json:"local_path,omitempty"`
	CreatedAt    int64   `json:"created_at"`
	UpdatedAt    int64   `json:"updated_at"`
}
