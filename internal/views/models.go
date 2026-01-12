package views

import (
	"time"
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
	Status      string // Mapped from state
	Summary     string
	AgentOutput  string // Final message from agent
	GitDiff     string // Git diff of changes
	Logs        []UILogEntry
	CreatedAt   time.Time
	PreviewURL  string
}

type UILogEntry struct {
	Timestamp time.Time
	Message   string
	Type      string
}

// FeedData is the data for the feed page
type FeedData struct {
	Reviews  []*UITask
	Active   []*UITask
	Done     []*UITask
	Projects map[string]UIProject
	Filter   string
}
