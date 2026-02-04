package codex

import (
	"encoding/json"
	"fmt"
)

type ThreadItem interface {
	ItemType() string
}

type CommandExecutionStatus string

const (
	CommandInProgress CommandExecutionStatus = "in_progress"
	CommandCompleted  CommandExecutionStatus = "completed"
	CommandFailed     CommandExecutionStatus = "failed"
)

type CommandExecutionItem struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Command          string                 `json:"command"`
	AggregatedOutput string                 `json:"aggregated_output"`
	ExitCode         *int                   `json:"exit_code,omitempty"`
	Status           CommandExecutionStatus `json:"status"`
}

func (CommandExecutionItem) ItemType() string { return "command_execution" }

type PatchChangeKind string

const (
	PatchAdd    PatchChangeKind = "add"
	PatchDelete PatchChangeKind = "delete"
	PatchUpdate PatchChangeKind = "update"
)

type FileUpdateChange struct {
	Path string          `json:"path"`
	Kind PatchChangeKind `json:"kind"`
}

type PatchApplyStatus string

const (
	PatchCompleted PatchApplyStatus = "completed"
	PatchFailed    PatchApplyStatus = "failed"
)

type FileChangeItem struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"`
	Changes []FileUpdateChange `json:"changes"`
	Status  PatchApplyStatus   `json:"status"`
}

func (FileChangeItem) ItemType() string { return "file_change" }

type McpToolCallStatus string

const (
	McpInProgress McpToolCallStatus = "in_progress"
	McpCompleted  McpToolCallStatus = "completed"
	McpFailed     McpToolCallStatus = "failed"
)

type McpToolCallResult struct {
	Content           []map[string]any `json:"content"`
	StructuredContent any              `json:"structured_content"`
}

type McpToolCallError struct {
	Message string `json:"message"`
}

type McpToolCallItem struct {
	ID        string             `json:"id"`
	Type      string             `json:"type"`
	Server    string             `json:"server"`
	Tool      string             `json:"tool"`
	Arguments any                `json:"arguments"`
	Result    *McpToolCallResult `json:"result,omitempty"`
	Error     *McpToolCallError  `json:"error,omitempty"`
	Status    McpToolCallStatus  `json:"status"`
}

func (McpToolCallItem) ItemType() string { return "mcp_tool_call" }

type AgentMessageItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text"`
}

func (AgentMessageItem) ItemType() string { return "agent_message" }

type ReasoningItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text"`
}

func (ReasoningItem) ItemType() string { return "reasoning" }

type WebSearchItem struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Query string `json:"query"`
}

func (WebSearchItem) ItemType() string { return "web_search" }

type ErrorItem struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (ErrorItem) ItemType() string { return "error" }

type TodoItem struct {
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

type TodoListItem struct {
	ID    string     `json:"id"`
	Type  string     `json:"type"`
	Items []TodoItem `json:"items"`
}

func (TodoListItem) ItemType() string { return "todo_list" }

// ParseThreadItem parses a raw item into a typed ThreadItem.
func ParseThreadItem(raw map[string]any) (ThreadItem, error) {
	itemType, _ := raw["type"].(string)
	if itemType == "" {
		return nil, fmt.Errorf("missing item type")
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	switch itemType {
	case "command_execution":
		var item CommandExecutionItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return item, nil
	case "file_change":
		var item FileChangeItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return item, nil
	case "mcp_tool_call":
		var item McpToolCallItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return item, nil
	case "agent_message":
		var item AgentMessageItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return item, nil
	case "reasoning":
		var item ReasoningItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return item, nil
	case "web_search":
		var item WebSearchItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return item, nil
	case "todo_list":
		var item TodoListItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return item, nil
	case "error":
		var item ErrorItem
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		return item, nil
	default:
		return nil, fmt.Errorf("unsupported item type %q", itemType)
	}
}
