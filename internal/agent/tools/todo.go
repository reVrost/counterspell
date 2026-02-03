package tools

import (
	_ "embed"
	"fmt"
	"sync"
)

//go:embed todo.md
var todoDescription string

// TodoStatus represents the status of a todo item
type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
)

// TodoItem represents a single todo item
type TodoItem struct {
	Content    string     `json:"content"`
	Status     TodoStatus `json:"status"`
	ActiveForm string     `json:"active_form"`
}

// TodoState manages the todo list state
type TodoState struct {
	mu    sync.RWMutex
	todos []TodoItem
}

// NewTodoState creates a new todo state
func NewTodoState() *TodoState {
	return &TodoState{
		todos: []TodoItem{},
	}
}

// GetTodos returns a copy of the current todos
func (ts *TodoState) GetTodos() []TodoItem {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	result := make([]TodoItem, len(ts.todos))
	copy(result, ts.todos)
	return result
}

// SetTodos replaces the entire todo list
func (ts *TodoState) SetTodos(todos []TodoItem) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.todos = todos
}

// GetInProgressTask returns the currently active task's active_form, or empty string
func (ts *TodoState) GetInProgressTask() string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	for _, t := range ts.todos {
		if t.Status == TodoStatusInProgress {
			if t.ActiveForm != "" {
				return t.ActiveForm
			}
			return t.Content
		}
	}
	return ""
}

// GetProgress returns completed count and total count
func (ts *TodoState) GetProgress() (completed, total int) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	total = len(ts.todos)
	for _, t := range ts.todos {
		if t.Status == TodoStatusCompleted {
			completed++
		}
	}
	return
}

func (r *Registry) makeTodoTool() Tool {
	return Tool{
		Description: todoDescription,
		// Description: "Manage a structured task list for tracking progress on complex tasks. " +
		// 	"Use this for multi-step tasks, complex work requiring planning, or when the user provides multiple tasks. " +
		// 	"Each task needs content (imperative), status (pending/in_progress/completed), and active_form (present continuous).",
		Schema: map[string]any{
			"todos": map[string]any{
				"type":        "array",
				"description": "The updated todo list",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"content": map[string]any{
							"type":        "string",
							"description": "What needs to be done (imperative form, e.g., 'Add user authentication')",
						},
						"status": map[string]any{
							"type":        "string",
							"enum":        []string{"pending", "in_progress", "completed"},
							"description": "Task status: pending, in_progress, or completed",
						},
						"active_form": map[string]any{
							"type":        "string",
							"description": "Present continuous form shown during execution (e.g., 'Adding user authentication')",
						},
					},
					"required": []string{"content", "status", "active_form"},
				},
			},
		},
		Func: r.toolTodos,
	}
}

func (r *Registry) toolTodos(args map[string]any) string {
	todosRaw, ok := args["todos"]
	if !ok {
		return "error: todos parameter required"
	}

	todosArr, ok := todosRaw.([]any)
	if !ok {
		return "error: todos must be an array"
	}

	var newTodos []TodoItem
	for _, item := range todosArr {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return "error: each todo must be an object"
		}

		content, _ := itemMap["content"].(string)
		status, _ := itemMap["status"].(string)
		activeForm, _ := itemMap["active_form"].(string)

		if content == "" {
			return "error: todo content is required"
		}

		var todoStatus TodoStatus
		switch status {
		case "pending":
			todoStatus = TodoStatusPending
		case "in_progress":
			todoStatus = TodoStatusInProgress
		case "completed":
			todoStatus = TodoStatusCompleted
		default:
			return fmt.Sprintf("error: invalid status %q, must be pending/in_progress/completed", status)
		}

		newTodos = append(newTodos, TodoItem{
			Content:    content,
			Status:     todoStatus,
			ActiveForm: activeForm,
		})
	}

	if r.ctx.TodoState == nil {
		return "error: todo state not initialized"
	}

	// Track what changed for the response
	oldTodos := r.ctx.TodoState.GetTodos()
	oldStatusMap := make(map[string]TodoStatus)
	for _, t := range oldTodos {
		oldStatusMap[t.Content] = t.Status
	}

	// Update the state
	r.ctx.TodoState.SetTodos(newTodos)

	// Emit todo update event
	if r.ctx.TodoEvents != nil {
		r.ctx.TodoEvents <- r.ctx.TodoState.GetTodos()
	}

	// Calculate stats
	var pending, inProgress, completed int
	var justStarted string
	var justCompleted []string

	for _, t := range newTodos {
		switch t.Status {
		case TodoStatusPending:
			pending++
		case TodoStatusInProgress:
			inProgress++
			oldStatus, existed := oldStatusMap[t.Content]
			if !existed || oldStatus != TodoStatusInProgress {
				if t.ActiveForm != "" {
					justStarted = t.ActiveForm
				} else {
					justStarted = t.Content
				}
			}
		case TodoStatusCompleted:
			completed++
			oldStatus, existed := oldStatusMap[t.Content]
			if existed && oldStatus != TodoStatusCompleted {
				justCompleted = append(justCompleted, t.Content)
			}
		}
	}

	response := "Todo list updated. "
	response += fmt.Sprintf("Status: %d pending, %d in progress, %d completed. ", pending, inProgress, completed)

	if justStarted != "" {
		response += fmt.Sprintf("Started: %s. ", justStarted)
	}
	if len(justCompleted) > 0 {
		response += fmt.Sprintf("Completed: %v. ", justCompleted)
	}

	response += "Continue with current tasks."

	return response
}
