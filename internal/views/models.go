package views

import (
	"time"

	"github.com/revrost/code/counterspell/internal/models"
)

// UIProject represents a project for display
type UIProject struct {
	ID    string
	Name  string
	Icon  string
	Color string
}

// UITask represents a task for display
type UITask struct {
	ID          string // Using string ID as per internal models
	ProjectID   string
	Description string // Mapped from Title
	AgentName   string
	Progress    int
	Status      models.TaskStatus
	Summary     string
	AgentOutput string // Final message from agent
	GitDiff     string // Git diff of changes
	Logs        []UILogEntry
	Messages    []UIMessage // Conversation history
	CreatedAt   time.Time
	PreviewURL  string
}

// UIMessage represents a message in the conversation history
type UIMessage struct {
	Role    string        // "user" or "assistant"
	Content []UIContent   // Content blocks
}

// UIContent represents a content block in a message
type UIContent struct {
	Type      string // "text", "tool_use", "tool_result"
	Text      string // For text blocks
	ToolName  string // For tool_use blocks
	ToolInput string // JSON string of tool input
	ToolID    string // Tool use ID
}

type UILogEntry struct {
	Timestamp time.Time
	Message   string
	Type      string
}

// FeedData is the data for the feed page
type FeedData struct {
	Todo     []*UITask
	Reviews  []*UITask
	Active   []*UITask
	Done     []*UITask
	Projects map[string]UIProject
	Filter   string
}

// UITodo represents a todo item for display
type UITodo struct {
	Content    string `json:"content"`
	Status     string `json:"status"` // pending, in_progress, completed
	ActiveForm string `json:"active_form"`
}
