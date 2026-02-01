// Package agent implements a simple coding agent loop.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/revrost/counterspell/internal/agent/tools"
	"github.com/revrost/counterspell/internal/llm"
)

// Event types for streaming
const (
	EventTool       = "tool"
	EventToolResult = "result"
	EventText       = "text"
	EventUserText   = "user_text"
	EventError      = "error"
	EventDone       = "done"
	EventTodo       = "todo" // Todo list update
)

// Message represents a single message in the conversation.
// The agent appends messages with both text and tool calls.
type Message struct {
	Role    string         `json:"role"` // "user" or "assistant"
	Content []ContentBlock `json:"content"`
}

// ContentBlock can be one of three types:
// - text: Assistant response text
// - tool_use: Assistant requesting to call a tool
// - tool_result: Result sent back after tool execution (from us to API)
type ContentBlock struct {
	// For all blocks
	Type string `json:"type"`

	// For text blocks
	Text string `json:"text,omitempty"`

	// For tool_use blocks (assistant calling a tool)
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`
	ID    string         `json:"id,omitempty"` // Links to tool_result

	// For tool_result blocks (result sent back to assistant)
	ToolUseID string `json:"tool_use_id,omitempty"` // Links to tool_use ID
	Content   string `json:"content,omitempty"`     // Tool output
}

// StreamEvent represents a single event in the agent execution.
type StreamEvent struct {
	Type       string `json:"type"`
	Content    string `json:"content"`
	Tool       string `json:"tool,omitempty"`
	Args       string `json:"args,omitempty"`
	SessionID  string `json:"session_id,omitempty"` // Backend session ID (e.g., Claude Code session)
}

// StreamCallback is called for each event during agent execution.
type StreamCallback func(event StreamEvent)

// Runner executes agent tasks with streaming output.
type Runner struct {
	provider       llm.Provider
	llmCaller      LLMCaller
	workDir        string
	callback       StreamCallback
	systemPrompt   string
	finalMessage   string
	messageHistory []Message
	todoState      *tools.TodoState
	toolRegistry   *tools.Registry
}

// NewRunner creates a new agent runner.
func NewRunner(provider llm.Provider, workDir string, callback StreamCallback) *Runner {
	r := &Runner{
		provider:     provider,
		llmCaller:    NewLLMCaller(provider),
		workDir:      workDir,
		callback:     callback,
		systemPrompt: fmt.Sprintf("You are a coding assistant. Work directory: %s. Be concise. Make changes directly.", workDir),
		todoState:    tools.NewTodoState(),
	}

	// Create tool registry with context
	toolCtx := &tools.Context{
		WorkDir:      workDir,
		TodoState:    r.todoState,
		OnTodoUpdate: r.emitTodoUpdate,
	}
	r.toolRegistry = tools.NewRegistry(toolCtx)

	return r
}

// GetFinalMessage returns the accumulated final message from the agent.
func (r *Runner) GetFinalMessage() string {
	return r.finalMessage
}

// GetMessageHistory returns the message history as JSON string.
func (r *Runner) GetMessageHistory() string {
	data, err := json.Marshal(r.messageHistory)
	if err != nil {
		return ""
	}
	return string(data)
}

// SetMessageHistory sets the initial message history from JSON string.
func (r *Runner) SetMessageHistory(historyJSON string) error {
	if historyJSON == "" {
		return nil
	}
	return json.Unmarshal([]byte(historyJSON), &r.messageHistory)
}

// GetTodos returns the current todo list as JSON string.
func (r *Runner) GetTodos() string {
	data, _ := json.Marshal(r.todoState.GetTodos())
	return string(data)
}

// GetTodoState returns the todo state.
func (r *Runner) GetTodoState() *tools.TodoState {
	return r.todoState
}

// Run executes the agent loop for a given task.
func (r *Runner) Run(ctx context.Context, task string) error {
	return r.runWithMessage(ctx, task, false)
}

// Continue resumes the agent loop with a new follow-up message.
// func (r *Runner) Continue(ctx context.Context, followUpMessage string) error {
// 	return r.runWithMessage(ctx, followUpMessage, true)
// }

// runWithMessage is the core loop that handles both new runs and continuations.
func (r *Runner) runWithMessage(ctx context.Context, userMessage string, isContinuation bool) error {
	allTools := r.toolRegistry.All()

	// Use existing message history or start fresh
	messages := r.messageHistory
	if messages == nil {
		messages = []Message{}
	}

	if userMessage != "" {
		// Add user message
		msg := Message{
			Role: "user",
			Content: []ContentBlock{
				{Type: "text", Text: userMessage},
			},
		}
		messages = append(messages, msg)

		r.emit(StreamEvent{Type: EventUserText, Content: userMessage})
	} else if len(messages) == 0 {
		return fmt.Errorf("agent: cannot start task with empty message")
	}

	// Agent loop
	for {
		select {
		case <-ctx.Done():
			r.messageHistory = messages
			return ctx.Err()
		default:
		}

		slog.Info("[RUNNER] Calling LLM API", "messages", messages, "all_tools", allTools, "system_prompt", r.systemPrompt)
		resp, err := r.llmCaller.Call(messages, allTools, r.systemPrompt)
		if err != nil {
			r.messageHistory = messages
			r.emit(StreamEvent{Type: EventError, Content: err.Error()})
			slog.Error("[RUNNER] LLM API call failed", "error", err)
			return err
		}

		// Log the raw response for debugging
		respJSON, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Printf("\n=== LLM API RESPONSE ===\n%s\n=== END RESPONSE ===\n\n", string(respJSON))
		for i, block := range resp.Content {
			fmt.Printf("Block %d: type=%s, has_text=%v, has_name=%v, id=%s\n", i, block.Type, block.Text != "", block.Name != "", block.ID)
		}

		toolResults := []ContentBlock{}

		// Immediately add assistant message and emit so UI shows response right away
		assistantMsg := Message{Role: "assistant", Content: resp.Content}
		messages = append(messages, assistantMsg)

		for _, block := range resp.Content {
			if block.Type == "text" && block.Text != "" {
				r.emit(StreamEvent{Type: EventText, Content: block.Text})
				r.finalMessage += block.Text
			}

			if block.Type == "tool_use" {
				argsJSON, _ := json.Marshal(block.Input)
				r.emit(StreamEvent{
					Type:    EventTool,
					Tool:    block.Name,
					Args:    string(argsJSON),
					Content: fmt.Sprintf("Running %s", block.Name),
				})

				result := r.runTool(block.Name, block.Input, allTools)

				// Truncate result for display
				displayResult := result
				if len(displayResult) > 200 {
					displayResult = displayResult[:200] + "..."
				}
				r.emit(StreamEvent{
					Type:    EventToolResult,
					Tool:    block.Name,
					Content: displayResult,
				})

				toolResults = append(toolResults, ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   result,
				})
			}
		}

		if len(toolResults) == 0 {
			slog.Info("[RUNNER] No more tools to run, completing task")
			break
		}

		slog.Info("[RUNNER] Running %d tool result(s) through agent loop", "len_tool_results", len(toolResults))
		toolResultMsg := Message{Role: "tool_result", Content: toolResults}
		messages = append(messages, toolResultMsg)
	}

	// Store message history for future continuations
	r.messageHistory = messages

	r.emit(StreamEvent{Type: EventDone, Content: "Task completed"})
	return nil
}

func (r *Runner) emit(event StreamEvent) {
	if r.callback != nil {
		r.callback(event)
	}
}

func (r *Runner) runTool(name string, args map[string]any, allTools map[string]tools.Tool) string {
	tool, ok := allTools[name]
	if !ok {
		return fmt.Sprintf("error: unknown tool %s", name)
	}
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Printf("error in %s: %v\n", name, rec)
		}
	}()
	return tool.Func(args)
}

// emitTodoUpdate sends a todo update event through the callback
func (r *Runner) emitTodoUpdate() {
	if r.callback == nil {
		return
	}

	todos := r.todoState.GetTodos()
	data, _ := json.Marshal(todos)

	r.emit(StreamEvent{
		Type:    EventTodo,
		Content: string(data),
	})
}
