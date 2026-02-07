package tools

import "fmt"

// makeMarkDoneTool creates a tool that marks the task as successfully completed.
func (r *Registry) makeMarkDoneTool() Tool {
	return Tool{
		Description: "Mark the task as successfully completed. Call this when you have finished the requested operation (e.g., successfully merged changes to main). This will update the task status to 'done'.",
		Schema:      map[string]any{},
		Func: func(args map[string]any) string {
			if r.ctx.TaskDone == nil {
				return "error: TaskDone callback not set"
			}

			if err := r.ctx.TaskDone(); err != nil {
				return fmt.Sprintf("error: failed to mark task as done: %v", err)
			}

			return "Task successfully marked as done. The task status has been updated to 'done'."
		},
	}
}
