package models

import "time"

// Agent represents a system-wide agent configuration.
type Agent struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	SystemPrompt string    `json:"system_prompt"`
	Tools        []string  `json:"tools"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ValidTools is the list of tools an agent can be assigned.
var ValidTools = []string{
	"bash",
	"edit",
	"glob",
	"grep",
	"ls",
	"multiedit",
	"read",
	"todo",
	"write",
}
