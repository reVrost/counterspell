package services

import (
	"context"
	"fmt"
	"html"
	"log/slog"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/revrost/code/counterspell/internal/git"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/pkg/agent"
)

// TaskResult represents the result of a completed task.
type TaskResult struct {
	TaskID      string
	Success     bool
	AgentOutput string
	GitDiff     string
	Error       string
}

// TaskJob represents a job submitted to the worker pool.
type TaskJob struct {
	TaskID    string
	ProjectID string
	Intent    string
	Owner     string
	Repo      string
	Token     string
	ResultCh  chan<- TaskResult
}

// Orchestrator manages task execution with agents.
type Orchestrator struct {
	tasks       *TaskService
	github      *GitHubService
	events      *EventBus
	settings    *SettingsService
	repos       *git.RepoManager
	dataDir     string
	workerPool  *ants.Pool
	resultCh    chan TaskResult
	running     map[string]context.CancelFunc
	mu          sync.Mutex
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(
	tasks *TaskService,
	github *GitHubService,
	events *EventBus,
	settings *SettingsService,
	dataDir string,
) (*Orchestrator, error) {
	// Create worker pool with 5 workers
	pool, err := ants.NewPool(5, ants.WithPreAlloc(false))
	if err != nil {
		return nil, fmt.Errorf("failed to create worker pool: %w", err)
	}

	orch := &Orchestrator{
		tasks:       tasks,
		github:      github,
		events:      events,
		settings:    settings,
		repos:       git.NewRepoManager(dataDir),
		dataDir:     dataDir,
		workerPool:  pool,
		resultCh:    make(chan TaskResult, 100),
		running:     make(map[string]context.CancelFunc),
	}

	// Start result processor goroutine
	go orch.processResults()

	return orch, nil
}

// Shutdown gracefully shuts down the orchestrator.
func (o *Orchestrator) Shutdown() {
	o.workerPool.Release()
	close(o.resultCh)
}

// StartTask creates a task and begins execution.
func (o *Orchestrator) StartTask(ctx context.Context, projectID, intent string) (*models.Task, error) {
	// First, verify project exists
	projects, err := o.github.GetProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	var owner, repo string
	projectExists := false
	for _, p := range projects {
		if p.ID == projectID {
			owner = p.GitHubOwner
			repo = p.GitHubRepo
			projectExists = true
			break
		}
	}

	if !projectExists {
		slog.Error("Project ID not found in database", "project_id", projectID, "total_projects", len(projects))
		return nil, fmt.Errorf("project not found (ID: %s). Please select a valid project.", projectID)
	}

	// Get GitHub token
	conn, err := o.github.GetActiveConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("no GitHub connection: %w", err)
	}

	// Create task in DB (after validating project exists)
	task, err := o.tasks.Create(ctx, projectID, intent, intent)
	if err != nil {
		slog.Error("Failed to create task in database", "project_id", projectID, "error", err)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	slog.Info("Task created successfully", "task_id", task.ID, "project_id", projectID, "owner", owner, "repo", repo)

	// Submit job to worker pool
	job := &TaskJob{
		TaskID:    task.ID,
		ProjectID: projectID,
		Intent:    intent,
		Owner:     owner,
		Repo:      repo,
		Token:     conn.Token,
		ResultCh:  o.resultCh,
	}

	if err := o.workerPool.Submit(func() {
		o.executeTask(job)
	}); err != nil {
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}

	return task, nil
}

// executeTask runs the agent loop for a task in a worker.
func (o *Orchestrator) executeTask(job *TaskJob) {
	ctx, cancel := context.WithCancel(context.Background())

	o.mu.Lock()
	o.running[job.TaskID] = cancel
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		delete(o.running, job.TaskID)
		o.mu.Unlock()
	}()

	// Update status to in_progress
	if err := o.tasks.UpdateStatus(ctx, job.TaskID, models.StatusInProgress); err != nil {
		o.emitError(job.TaskID, "Failed to update status")
		job.ResultCh <- TaskResult{
			TaskID: job.TaskID,
			Success: false,
			Error:   "Failed to update status",
		}
		return
	}

	o.emit(job.TaskID, "plan", "Starting task execution...")
	o.emit(job.TaskID, "info", fmt.Sprintf("Preparing %s/%s...", job.Owner, job.Repo))

	// Clone/fetch repo
	repoPath, err := o.repos.EnsureRepo(job.Owner, job.Repo, job.Token)
	if err != nil {
		o.emitError(job.TaskID, "Failed to prepare repo: "+err.Error())
		job.ResultCh <- TaskResult{
			TaskID: job.TaskID,
			Success: false,
			Error:   "Failed to prepare repo: " + err.Error(),
		}
		return
	}

	o.emit(job.TaskID, "info", "Creating isolated workspace...")

	// Create worktree
	branchName := fmt.Sprintf("agent/task-%s", job.TaskID[:8])
	worktreePath, err := o.repos.CreateWorktree(job.Owner, job.Repo, job.TaskID, branchName)
	if err != nil {
		o.emitError(job.TaskID, "Failed to create worktree: "+err.Error())
		job.ResultCh <- TaskResult{
			TaskID: job.TaskID,
			Success: false,
			Error:   "Failed to create worktree: " + err.Error(),
		}
		return
	}

	o.emit(job.TaskID, "success", fmt.Sprintf("Workspace ready: %s", branchName))

	// Get API key
	settings, err := o.settings.GetSettings(ctx)
	if err != nil || settings.AnthropicKey == "" {
		o.emitError(job.TaskID, "No Anthropic API key configured. Add it in Settings.")
		job.ResultCh <- TaskResult{
			TaskID: job.TaskID,
			Success: false,
			Error:   "No Anthropic API key configured",
		}
		return
	}

	o.emit(job.TaskID, "plan", "Starting agent...")

	// Create agent runner with streaming callback
	runner := agent.NewRunner(settings.AnthropicKey, worktreePath, func(event agent.StreamEvent) {
		o.handleAgentEvent(job.TaskID, event)
	})

	// Run the agent
	if err := runner.Run(ctx, job.Intent); err != nil {
		if ctx.Err() != nil {
			o.emit(job.TaskID, "info", "Task cancelled")
		} else {
			o.emitError(job.TaskID, "Agent failed: "+err.Error())
		}
		job.ResultCh <- TaskResult{
			TaskID:  job.TaskID,
			Success: false,
			Error:   "Agent failed: " + err.Error(),
		}
		return
	}

	// Commit and push changes
	o.emit(job.TaskID, "info", "Committing changes...")

	commitMsg := fmt.Sprintf("feat: %s\n\nTask ID: %s", job.Intent, job.TaskID)
	if err := o.repos.CommitAndPush(job.TaskID, commitMsg); err != nil {
		o.emit(job.TaskID, "info", "No changes to commit or push failed: "+err.Error())
	} else {
		o.emit(job.TaskID, "success", fmt.Sprintf("Pushed to branch: %s", branchName))
	}

	// Get git diff
	gitDiff, err := o.repos.GetDiff(job.TaskID)
	if err != nil {
		slog.Warn("Failed to get git diff", "task_id", job.TaskID, "error", err)
		gitDiff = ""
	}

	// Get final agent output
	agentOutput := runner.GetFinalMessage()

	// Send result to channel for processing
	job.ResultCh <- TaskResult{
		TaskID:      job.TaskID,
		Success:     true,
		AgentOutput: agentOutput,
		GitDiff:     gitDiff,
	}

	slog.Info("Task completed", "task_id", job.TaskID, "repo", fmt.Sprintf("%s/%s", job.Owner, job.Repo), "path", repoPath)
}

// processResults processes task results from the channel.
// This runs in a single goroutine to ensure SQLite concurrency safety.
func (o *Orchestrator) processResults() {
	for result := range o.resultCh {
		ctx := context.Background()

		// Update task with result
		status := models.StatusReview
		if !result.Success {
			status = models.StatusReview // Keep review status so user can see the error
		}

		if err := o.tasks.UpdateWithResult(ctx, result.TaskID, status, result.AgentOutput, result.GitDiff); err != nil {
			slog.Error("Failed to update task with result", "task_id", result.TaskID, "error", err)
			continue
		}

		if result.Success {
			o.emit(result.TaskID, "success", "Task complete - ready for review")
		} else {
			o.emitError(result.TaskID, result.Error)
		}
	}
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
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
