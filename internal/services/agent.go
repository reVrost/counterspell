package services

import (
	"context"
	"time"

	"github.com/revrost/code/counterspell/internal/git"
	"github.com/revrost/code/counterspell/internal/models"
)

// AgentRunner executes agent tasks in isolated worktrees.
type AgentRunner struct {
	taskService *TaskService
	eventBus    *EventBus
	worktreeMgr *git.WorktreeManager
}

// NewAgentRunner creates a new agent runner.
func NewAgentRunner(taskService *TaskService, eventBus *EventBus, worktreeMgr *git.WorktreeManager) *AgentRunner {
	return &AgentRunner{
		taskService: taskService,
		eventBus:    eventBus,
		worktreeMgr: worktreeMgr,
	}
}

// Run executes a task with an agent.
func (r *AgentRunner) Run(ctx context.Context, taskID string) error {
	task, err := r.taskService.Get(ctx, taskID)
	if err != nil {
		r.eventBus.Publish(models.Event{
			TaskID:      taskID,
			Type:        "error",
			HTMLPayload: "Failed to get task",
		})
		return err
	}

	// Publish planning log
	r.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        "log",
		HTMLPayload: `<span class="text-yellow-400">[plan]</span> Analyzing task: ` + task.Title,
	})

	// Create worktree for isolated execution
	worktreePath, err := r.worktreeMgr.CreateWorktree(taskID, "main")
	if err != nil {
		r.eventBus.Publish(models.Event{
			TaskID:      taskID,
			Type:        "error",
			HTMLPayload: `<span class="text-red-400">[error]</span> Failed to create worktree: ` + err.Error(),
		})
		return err
	}

	r.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        "log",
		HTMLPayload: `<span class="text-blue-400">[info]</span> Created worktree: ` + worktreePath,
	})

	time.Sleep(500 * time.Millisecond)

	// Simulate agent thinking
	r.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        "log",
		HTMLPayload: `<span class="text-purple-400">[code]</span> Writing code...`,
	})

	time.Sleep(2 * time.Second)

	// Simulate code completion
	r.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        "log",
		HTMLPayload: `<span class="text-green-400">[success]</span> Code written successfully`,
	})

	// Move to review
	if err := r.taskService.UpdateStatus(ctx, taskID, models.StatusReview); err != nil {
		r.eventBus.Publish(models.Event{
			TaskID:      taskID,
			Type:        "error",
			HTMLPayload: `<span class="text-red-400">[error]</span> Failed to update status`,
		})
		return err
	}

	r.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        "status",
		HTMLPayload: `<span class="text-green-400">✓</span> Moved to review`,
	})

	// Cleanup worktree (defer until approved or rejected)
	// For now, we keep it for review

	return nil
}
