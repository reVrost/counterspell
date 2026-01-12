package services

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"sync"

	"github.com/revrost/code/counterspell/internal/git"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/pkg/agent"
)

// Orchestrator manages task execution with agents.
type Orchestrator struct {
	tasks    *TaskService
	github   *GitHubService
	events   *EventBus
	settings *SettingsService
	repos    *git.RepoManager
	dataDir  string

	// Track running tasks
	running map[string]context.CancelFunc
	mu      sync.Mutex
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(
	tasks *TaskService,
	github *GitHubService,
	events *EventBus,
	settings *SettingsService,
	dataDir string,
) *Orchestrator {
	return &Orchestrator{
		tasks:    tasks,
		github:   github,
		events:   events,
		settings: settings,
		repos:    git.NewRepoManager(dataDir),
		dataDir:  dataDir,
		running:  make(map[string]context.CancelFunc),
	}
}

// StartTask creates a task and begins execution.
func (o *Orchestrator) StartTask(ctx context.Context, projectID, intent string) (*models.Task, error) {
	// Create task in DB
	task, err := o.tasks.Create(ctx, projectID, intent, intent)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Start execution in background
	go o.executeTask(task.ID, projectID, intent)

	return task, nil
}

// executeTask runs the agent loop for a task.
func (o *Orchestrator) executeTask(taskID, projectID, intent string) {
	ctx, cancel := context.WithCancel(context.Background())

	o.mu.Lock()
	o.running[taskID] = cancel
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		delete(o.running, taskID)
		o.mu.Unlock()
	}()

	// Update status to in_progress
	if err := o.tasks.UpdateStatus(ctx, taskID, models.StatusInProgress); err != nil {
		o.emitError(taskID, "Failed to update status")
		return
	}

	o.emit(taskID, "plan", "Starting task execution...")

	// Get project details
	projects, err := o.github.GetProjects(ctx)
	if err != nil {
		o.emitError(taskID, "Failed to get projects: "+err.Error())
		return
	}

	var owner, repo string
	for _, p := range projects {
		if p.ID == projectID {
			owner = p.GitHubOwner
			repo = p.GitHubRepo
			break
		}
	}

	if owner == "" || repo == "" {
		o.emitError(taskID, "Project not found")
		return
	}

	// Get GitHub token
	conn, err := o.github.GetActiveConnection(ctx)
	if err != nil {
		o.emitError(taskID, "No GitHub connection: "+err.Error())
		return
	}

	o.emit(taskID, "info", fmt.Sprintf("Preparing %s/%s...", owner, repo))

	// Clone/fetch repo
	repoPath, err := o.repos.EnsureRepo(owner, repo, conn.Token)
	if err != nil {
		o.emitError(taskID, "Failed to prepare repo: "+err.Error())
		return
	}

	o.emit(taskID, "info", "Creating isolated workspace...")

	// Create worktree
	branchName := fmt.Sprintf("agent/task-%s", taskID[:8])
	worktreePath, err := o.repos.CreateWorktree(owner, repo, taskID, branchName)
	if err != nil {
		o.emitError(taskID, "Failed to create worktree: "+err.Error())
		return
	}

	o.emit(taskID, "success", fmt.Sprintf("Workspace ready: %s", branchName))

	// Get API key
	settings, err := o.settings.GetSettings(ctx)
	if err != nil || settings.AnthropicKey == "" {
		o.emitError(taskID, "No Anthropic API key configured. Add it in Settings.")
		return
	}

	o.emit(taskID, "plan", "Starting agent...")

	// Create agent runner with streaming callback
	runner := agent.NewRunner(settings.AnthropicKey, worktreePath, func(event agent.StreamEvent) {
		o.handleAgentEvent(taskID, event)
	})

	// Run the agent
	if err := runner.Run(ctx, intent); err != nil {
		if ctx.Err() != nil {
			o.emit(taskID, "info", "Task cancelled")
		} else {
			o.emitError(taskID, "Agent failed: "+err.Error())
		}
		return
	}

	// Commit and push changes
	o.emit(taskID, "info", "Committing changes...")

	commitMsg := fmt.Sprintf("feat: %s\n\nTask ID: %s", intent, taskID)
	if err := o.repos.CommitAndPush(taskID, commitMsg); err != nil {
		o.emit(taskID, "info", "No changes to commit or push failed: "+err.Error())
	} else {
		o.emit(taskID, "success", fmt.Sprintf("Pushed to branch: %s", branchName))
	}

	// Update status to review
	if err := o.tasks.UpdateStatus(ctx, taskID, models.StatusReview); err != nil {
		o.emitError(taskID, "Failed to update status")
		return
	}

	o.emit(taskID, "success", "Task complete - ready for review")
	slog.Info("Task completed", "task_id", taskID, "repo", fmt.Sprintf("%s/%s", owner, repo), "path", repoPath)
}

// handleAgentEvent converts agent events to UI events.
func (o *Orchestrator) handleAgentEvent(taskID string, event agent.StreamEvent) {
	switch event.Type {
	case agent.EventPlan:
		o.emit(taskID, "plan", event.Content)
	case agent.EventTool:
		o.emit(taskID, "code", fmt.Sprintf("[%s] %s", event.Tool, truncate(event.Args, 60)))
	case agent.EventResult:
		o.emit(taskID, "info", truncate(event.Content, 100))
	case agent.EventText:
		o.emit(taskID, "text", event.Content)
	case agent.EventError:
		o.emitError(taskID, event.Content)
	case agent.EventDone:
		o.emit(taskID, "success", event.Content)
	}
}

// CancelTask cancels a running task.
func (o *Orchestrator) CancelTask(taskID string) {
	o.mu.Lock()
	cancel, ok := o.running[taskID]
	o.mu.Unlock()

	if ok {
		cancel()
	}
}

// emit sends an event to the UI.
func (o *Orchestrator) emit(taskID, level, message string) {
	colorClass := "text-gray-400"
	switch level {
	case "plan":
		colorClass = "text-yellow-400"
	case "code":
		colorClass = "text-purple-400"
	case "info":
		colorClass = "text-blue-400"
	case "success":
		colorClass = "text-green-400"
	case "error":
		colorClass = "text-red-400"
	}

	htmlPayload := fmt.Sprintf(`<span class="%s">[%s]</span> %s`, colorClass, level, html.EscapeString(message))

	o.events.Publish(models.Event{
		TaskID:      taskID,
		Type:        "log",
		HTMLPayload: htmlPayload,
	})
}

func (o *Orchestrator) emitError(taskID, message string) {
	o.emit(taskID, "error", message)
	// Also update task status to review so user can see the error
	o.tasks.UpdateStatus(context.Background(), taskID, models.StatusReview)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
