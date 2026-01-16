package services

import (
	"bufio"
	"context"
	"log/slog"
	"os/exec"
	"strings"

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
			Type:        models.EventTypeError,
			HTMLPayload: "Failed to get task",
		})
		return err
	}

	// Publish planning log
	r.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        models.EventTypeLog,
		HTMLPayload: `<span class="text-yellow-400">[plan]</span> Analyzing task: ` + task.Title,
	})

	// Create worktree for isolated execution
	worktreePath, err := r.worktreeMgr.CreateWorktree(taskID, "main")
	if err != nil {
		r.eventBus.Publish(models.Event{
			TaskID:      taskID,
			Type:        models.EventTypeError,
			HTMLPayload: `<span class="text-red-400">[error]</span> Failed to create worktree: ` + err.Error(),
		})
		return err
	}

	r.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        models.EventTypeLog,
		HTMLPayload: `<span class="text-blue-400">[info]</span> Created worktree: ` + worktreePath,
	})

	// Execute crush command with task's intent
	r.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        models.EventTypeLog,
		HTMLPayload: `<span class="text-purple-400">[code]</span> Starting agent...`,
	})

	cmd := exec.CommandContext(ctx, "crush", "run", "--yolo", task.Intent)
	cmd.Dir = worktreePath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.eventBus.Publish(models.Event{
			TaskID:      taskID,
			Type:        models.EventTypeError,
			HTMLPayload: `<span class="text-red-400">[error]</span> Failed to create stdout pipe: ` + err.Error(),
		})
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		r.eventBus.Publish(models.Event{
			TaskID:      taskID,
			Type:        models.EventTypeError,
			HTMLPayload: `<span class="text-red-400">[error]</span> Failed to create stderr pipe: ` + err.Error(),
		})
		return err
	}

	if err := cmd.Start(); err != nil {
		r.eventBus.Publish(models.Event{
			TaskID:      taskID,
			Type:        models.EventTypeError,
			HTMLPayload: `<span class="text-red-400">[error]</span> Failed to start agent: ` + err.Error(),
		})
		return err
	}

	// Stream output in real-time
	outputScanner := bufio.NewScanner(stdout)
	errorScanner := bufio.NewScanner(stderr)

	done := make(chan bool, 2)

	go func() {
		for outputScanner.Scan() {
			line := outputScanner.Text()
			r.eventBus.Publish(models.Event{
				TaskID:      taskID,
				Type:        models.EventTypeLog,
				HTMLPayload: `<span class="text-gray-300">` + escapeHTML(line) + `</span>`,
			})
		}
		done <- true
	}()

	go func() {
		for errorScanner.Scan() {
			line := errorScanner.Text()
			if strings.Contains(strings.ToLower(line), "error") {
				r.eventBus.Publish(models.Event{
					TaskID:      taskID,
					Type:        models.EventTypeLog,
					HTMLPayload: `<span class="text-red-400">` + escapeHTML(line) + `</span>`,
				})
			} else {
				r.eventBus.Publish(models.Event{
					TaskID:      taskID,
					Type:        models.EventTypeLog,
					HTMLPayload: `<span class="text-yellow-400">` + escapeHTML(line) + `</span>`,
				})
			}
		}
		done <- true
	}()

	// Wait for both scanners to finish
	<-done
	<-done

	// Wait for command to finish
	err = cmd.Wait()

	if err != nil {
		r.eventBus.Publish(models.Event{
			TaskID:      taskID,
			Type:        models.EventTypeError,
			HTMLPayload: `<span class="text-red-400">[error]</span> Agent failed: ` + err.Error(),
		})
		return err
	}

	// Move to review
	if err := r.taskService.UpdateStatus(ctx, taskID, models.StatusReview); err != nil {
		r.eventBus.Publish(models.Event{
			TaskID:      taskID,
			Type:        models.EventTypeError,
			HTMLPayload: `<span class="text-red-400">[error]</span> Failed to update status`,
		})
		return err
	}

	r.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        models.EventTypeStatus,
		HTMLPayload: `<span class="text-green-400">✓</span> Moved to review`,
	})

	slog.Info("Agent completed", "task_id", taskID, "title", task.Title)

	// Cleanup worktree (defer until approved or rejected)
	// For now, we keep it for review

	return nil
}

// escapeHTML escapes HTML special characters.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
