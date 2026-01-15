package services

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/revrost/code/counterspell/internal/git"
	"github.com/revrost/code/counterspell/internal/llm"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/agent"
)

// TaskResult represents the result of a completed task.
type TaskResult struct {
	TaskID         string
	Success        bool
	AgentOutput    string
	GitDiff        string
	MessageHistory string
	Error          string
}

// TaskJob represents a job submitted to the worker pool.
type TaskJob struct {
	TaskID         string
	ProjectID      string
	Intent         string
	ModelID        string
	Owner          string
	Repo           string
	Token          string
	MessageHistory string // For continuations
	IsContinuation bool
	ResultCh       chan<- TaskResult
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
	slog.Info("[ORCHESTRATOR] Creating new orchestrator", "data_dir", dataDir)
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

	slog.Info("[ORCHESTRATOR] Worker pool created", "workers", 5, "prealloc", false)

	// Start result processor goroutine
	go orch.processResults()

	slog.Info("[ORCHESTRATOR] Orchestrator initialized")
	return orch, nil
}

// Shutdown gracefully shuts down the orchestrator.
func (o *Orchestrator) Shutdown() {
	slog.Info("[ORCHESTRATOR] Shutting down")
	o.workerPool.Release()
	close(o.resultCh)
	slog.Info("[ORCHESTRATOR] Shutdown complete")
}

// StartTask creates a task and begins execution.
func (o *Orchestrator) StartTask(ctx context.Context, projectID, intent, modelID string) (*models.Task, error) {
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
	tokenLen := len(conn.Token)
	slog.Info("GitHub token retrieved", "token_length", tokenLen, "has_token", tokenLen > 0)
	if tokenLen == 0 {
		slog.Error("GitHub token is empty - need to configure in settings")
		return nil, fmt.Errorf("GitHub token is not configured. Please add a GitHub connection in Settings.")
	}

	// Create task in DB (after validating project exists)
	task, err := o.tasks.Create(ctx, projectID, intent, intent)
	if err != nil {
		slog.Error("Failed to create task in database", "project_id", projectID, "error", err)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	slog.Info("Task created successfully", "task_id", task.ID, "project_id", projectID, "owner", owner, "repo", repo, "model_id", modelID)

	// Emit task_created event to trigger feed refresh
	o.events.Publish(models.Event{
		TaskID: task.ID,
		Type:   "task_created",
	})

	slog.Info("Submitting task to worker pool", "task_id", task.ID, "intent", intent)

	// Submit job to worker pool
	job := &TaskJob{
		TaskID:    task.ID,
		ProjectID: projectID,
		Intent:    intent,
		ModelID:   modelID,
		Owner:     owner,
		Repo:      repo,
		Token:     conn.Token,
		ResultCh:  o.resultCh,
	}

	slog.Info("[ORCHESTRATOR] Worker pool state before submit", "task_id", task.ID, "running_workers", o.workerPool.Running(), "waiting_tasks", o.workerPool.Waiting())

	if err := o.workerPool.Submit(func() {
		slog.Info("[AGENT LOOP] Worker started for task", "task_id", job.TaskID)
		o.executeTask(job)
	}); err != nil {
		slog.Error("Failed to submit task to worker pool", "task_id", task.ID, "error", err)
		return nil, fmt.Errorf("failed to submit task: %w", err)
	}

	slog.Info("Task submitted to worker pool successfully", "task_id", task.ID, "running_workers", o.workerPool.Running(), "waiting_tasks", o.workerPool.Waiting())
	return task, nil
}

// executeTask runs the agent loop for a task in a worker.
func (o *Orchestrator) executeTask(job *TaskJob) {
	ctx, cancel := context.WithCancel(context.Background())

	slog.Info("[AGENT LOOP] executeTask started", "task_id", job.TaskID, "owner", job.Owner, "repo", job.Repo, "model_id", job.ModelID)

	o.mu.Lock()
	o.running[job.TaskID] = cancel
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		delete(o.running, job.TaskID)
		o.mu.Unlock()
		slog.Info("[AGENT LOOP] executeTask finished", "task_id", job.TaskID)
	}()

	// Update status to in_progress
	slog.Info("[AGENT LOOP] Updating task status to in_progress", "task_id", job.TaskID)
	if err := o.tasks.UpdateStatus(ctx, job.TaskID, models.StatusInProgress); err != nil {
		slog.Error("[AGENT LOOP] Failed to update task status", "task_id", job.TaskID, "error", err)
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
	slog.Info("[AGENT LOOP] Ensuring repo exists", "task_id", job.TaskID, "owner", job.Owner, "repo", job.Repo)
	repoPath, err := o.repos.EnsureRepo(job.Owner, job.Repo, job.Token)
	if err != nil {
		slog.Error("[AGENT LOOP] Failed to prepare repo", "task_id", job.TaskID, "error", err)
		o.emitError(job.TaskID, "Failed to prepare repo: "+err.Error())
		job.ResultCh <- TaskResult{
			TaskID: job.TaskID,
			Success: false,
			Error:   "Failed to prepare repo: " + err.Error(),
		}
		return
	}
	slog.Info("[AGENT LOOP] Repo ready", "task_id", job.TaskID, "repo_path", repoPath)

	o.emit(job.TaskID, "info", "Creating isolated workspace...")

	// Create worktree
	branchName := fmt.Sprintf("agent/task-%s", job.TaskID[:8])
	slog.Info("[AGENT LOOP] Creating worktree", "task_id", job.TaskID, "branch", branchName)
	worktreePath, err := o.repos.CreateWorktree(job.Owner, job.Repo, job.TaskID, branchName)
	if err != nil {
		slog.Error("[AGENT LOOP] Failed to create worktree", "task_id", job.TaskID, "error", err)
		o.emitError(job.TaskID, "Failed to create worktree: "+err.Error())
		job.ResultCh <- TaskResult{
			TaskID: job.TaskID,
			Success: false,
			Error:   "Failed to create worktree: " + err.Error(),
		}
		return
	}
	slog.Info("[AGENT LOOP] Worktree created successfully", "task_id", job.TaskID, "worktree_path", worktreePath)

	o.emit(job.TaskID, "success", fmt.Sprintf("Workspace ready: %s", branchName))

	// Get API key and create provider
	slog.Info("[AGENT LOOP] Getting settings", "task_id", job.TaskID)
	settings, err := o.settings.GetSettings(ctx)
	if err != nil {
		slog.Error("[AGENT LOOP] Failed to get settings", "task_id", job.TaskID, "error", err)
		o.emitError(job.TaskID, "Failed to get settings")
		job.ResultCh <- TaskResult{
			TaskID: job.TaskID,
			Success: false,
			Error:   "Failed to get settings",
		}
		return
	}
	slog.Info("[AGENT LOOP] Got settings", "task_id", job.TaskID,
		"openrouter_key_len", len(settings.OpenRouterKey),
		"zai_key_len", len(settings.ZaiKey),
		"anthropic_key_len", len(settings.AnthropicKey),
		"openai_key_len", len(settings.OpenAIKey))

	// Parse model_id to get provider and model name
	providerPrefix, modelName := llm.ParseModelID(job.ModelID)
	slog.Info("[AGENT LOOP] Parsed model ID", "task_id", job.TaskID, "provider", providerPrefix, "model", modelName)

	var provider llm.Provider
	switch providerPrefix {
	case "openrouter", "o":
		if settings.OpenRouterKey == "" {
			o.emitError(job.TaskID, "OpenRouter API key not configured")
			job.ResultCh <- TaskResult{
				TaskID: job.TaskID,
				Success: false,
				Error:   "OpenRouter API key not configured",
			}
			return
		}
		provider = llm.NewOpenRouterProvider(settings.OpenRouterKey)
	case "zai", "z":
		if settings.ZaiKey == "" {
			slog.Error("[AGENT LOOP] Z.ai API key not configured", "task_id", job.TaskID, "zai_key_len", len(settings.ZaiKey))
			o.emitError(job.TaskID, "Z.ai API key not configured")
			job.ResultCh <- TaskResult{
				TaskID: job.TaskID,
				Success: false,
				Error:   "Z.ai API key not configured",
			}
			return
		}
		slog.Info("[AGENT LOOP] Creating Z.ai provider", "task_id", job.TaskID, "zai_key_len", len(settings.ZaiKey))
		provider = llm.NewZaiProvider(settings.ZaiKey)
	case "anthropic":
		if settings.AnthropicKey == "" {
			o.emitError(job.TaskID, "Anthropic API key not configured")
			job.ResultCh <- TaskResult{
				TaskID: job.TaskID,
				Success: false,
				Error:   "Anthropic API key not configured",
			}
			return
		}
		provider = llm.NewAnthropicProvider(settings.AnthropicKey)
	default:
		// Try auto-detect based on available keys
		if settings.OpenRouterKey != "" {
			provider = llm.NewOpenRouterProvider(settings.OpenRouterKey)
		} else if settings.ZaiKey != "" {
			provider = llm.NewZaiProvider(settings.ZaiKey)
		} else if settings.AnthropicKey != "" {
			provider = llm.NewAnthropicProvider(settings.AnthropicKey)
		} else {
			o.emitError(job.TaskID, "No API key configured. Add OpenRouter, Z.ai, or Anthropic key in Settings.")
			job.ResultCh <- TaskResult{
				TaskID: job.TaskID,
				Success: false,
				Error:   "No API key configured",
			}
			return
		}
	}

	// Set model name on provider
	if modelName != "" {
		provider.SetModel(modelName)
	}

	finalModel := provider.Model()
	slog.Info("[AGENT LOOP] Provider configured", "task_id", job.TaskID, "final_model", finalModel)

	if job.IsContinuation {
		o.emit(job.TaskID, "plan", fmt.Sprintf("Continuing with model: %s...", finalModel))
	} else {
		o.emit(job.TaskID, "plan", fmt.Sprintf("Starting agent with model: %s...", finalModel))
	}

	// Create agent runner with streaming callback
	slog.Info("[AGENT LOOP] Creating agent runner", "task_id", job.TaskID, "worktree", worktreePath, "intent", job.Intent, "provider", providerPrefix, "model", finalModel, "is_continuation", job.IsContinuation)
	runner := agent.NewRunner(provider, worktreePath, func(event agent.StreamEvent) {
		slog.Debug("[AGENT LOOP] Agent event", "task_id", job.TaskID, "event_type", event.Type)
		o.handleAgentEvent(job.TaskID, event)
	})

	// Load message history for continuations
	if job.IsContinuation && job.MessageHistory != "" {
		if err := runner.SetMessageHistory(job.MessageHistory); err != nil {
			slog.Warn("[AGENT LOOP] Failed to load message history", "task_id", job.TaskID, "error", err)
		} else {
			slog.Info("[AGENT LOOP] Loaded message history", "task_id", job.TaskID)
		}
	}

	// Run the agent (Run for new tasks, Continue for continuations)
	slog.Info("[AGENT LOOP] Starting agent execution", "task_id", job.TaskID, "is_continuation", job.IsContinuation)
	var runErr error
	if job.IsContinuation {
		runErr = runner.Continue(ctx, job.Intent)
	} else {
		runErr = runner.Run(ctx, job.Intent)
	}
	if runErr != nil {
		slog.Error("[AGENT LOOP] Agent execution failed", "task_id", job.TaskID, "error", runErr)
		if ctx.Err() != nil {
			o.emit(job.TaskID, "info", "Task cancelled")
		} else {
			o.emitError(job.TaskID, "Agent failed: "+runErr.Error())
		}
		job.ResultCh <- TaskResult{
			TaskID:  job.TaskID,
			Success: false,
			Error:   "Agent failed: " + runErr.Error(),
		}
		return
	}
	slog.Info("[AGENT LOOP] Agent execution completed successfully", "task_id", job.TaskID)

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

	// Get final agent output and message history
	agentOutput := runner.GetFinalMessage()
	messageHistory := runner.GetMessageHistory()

	// Send result to channel for processing
	job.ResultCh <- TaskResult{
		TaskID:         job.TaskID,
		Success:        true,
		AgentOutput:    agentOutput,
		GitDiff:        gitDiff,
		MessageHistory: messageHistory,
	}

	slog.Info("Task completed", "task_id", job.TaskID, "repo", fmt.Sprintf("%s/%s", job.Owner, job.Repo), "path", repoPath)
}

// processResults processes task results from the channel.
// This runs in a single goroutine to ensure SQLite concurrency safety.
func (o *Orchestrator) processResults() {
	slog.Info("[ORCHESTRATOR] processResults goroutine started")
	for result := range o.resultCh {
		slog.Info("[ORCHESTRATOR] Received result from worker", "task_id", result.TaskID, "success", result.Success)
		ctx := context.Background()

		// Update task with result
		status := models.StatusReview
		if !result.Success {
			status = models.StatusReview // Keep review status so user can see the error
		}

		if err := o.tasks.UpdateWithResult(ctx, result.TaskID, status, result.AgentOutput, result.GitDiff, result.MessageHistory); err != nil {
			slog.Error("Failed to update task with result", "task_id", result.TaskID, "error", err)
			continue
		}

		if result.Success {
			o.emit(result.TaskID, "success", "Task complete - ready for review")
		} else {
			o.emitError(result.TaskID, result.Error)
		}

		// Emit status_change event to trigger SSE updates for diff/agent tabs
		o.events.Publish(models.Event{
			TaskID: result.TaskID,
			Type:   "status_change",
		})

		slog.Info("[ORCHESTRATOR] Result processed", "task_id", result.TaskID)
	}
	slog.Info("[ORCHESTRATOR] processResults goroutine ended")
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
	case agent.EventMessages:
		// Publish message history for live agent panel updates
		o.events.Publish(models.Event{
			TaskID:      taskID,
			Type:        "agent_update",
			HTMLPayload: event.Messages, // JSON message history
		})
	case agent.EventTodo:
		// Publish todo list for live todo panel updates
		o.events.Publish(models.Event{
			TaskID:      taskID,
			Type:        "todo",
			HTMLPayload: event.Content, // JSON todo list
		})
	}
}

// ContinueTask continues an existing task with a follow-up message.
func (o *Orchestrator) ContinueTask(ctx context.Context, taskID, followUpMessage, modelID string) error {
	// Get existing task
	task, err := o.tasks.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get project info
	projects, err := o.github.GetProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}

	var owner, repo string
	for _, p := range projects {
		if p.ID == task.ProjectID {
			owner = p.GitHubOwner
			repo = p.GitHubRepo
			break
		}
	}

	if owner == "" {
		return fmt.Errorf("project not found for task")
	}

	// Get GitHub token
	conn, err := o.github.GetActiveConnection(ctx)
	if err != nil {
		return fmt.Errorf("no GitHub connection: %w", err)
	}

	// Update task status back to in_progress
	if err := o.tasks.UpdateStatus(ctx, taskID, models.StatusInProgress); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Emit task_created event to trigger feed refresh
	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   "task_created",
	})

	// Immediately emit agent_update with user message appended so it shows right away
	updatedHistory := appendUserMessage(task.MessageHistory, followUpMessage)
	o.events.Publish(models.Event{
		TaskID:      taskID,
		Type:        "agent_update",
		HTMLPayload: updatedHistory,
	})

	slog.Info("Continuing task", "task_id", taskID, "follow_up", followUpMessage)

	// Submit job to worker pool
	job := &TaskJob{
		TaskID:         taskID,
		ProjectID:      task.ProjectID,
		Intent:         followUpMessage,
		ModelID:        modelID,
		Owner:          owner,
		Repo:           repo,
		Token:          conn.Token,
		MessageHistory: task.MessageHistory,
		IsContinuation: true,
		ResultCh:       o.resultCh,
	}

	if err := o.workerPool.Submit(func() {
		slog.Info("[AGENT LOOP] Worker started for continuation", "task_id", job.TaskID)
		o.executeTask(job)
	}); err != nil {
		return fmt.Errorf("failed to submit continuation: %w", err)
	}

	return nil
}

// MergeTask merges a task's branch to main and marks it done.
// If there's a merge conflict, returns ErrMergeConflict for the handler to show conflict UI.
func (o *Orchestrator) MergeTask(ctx context.Context, taskID string) error {
	// Get task to find project info
	task, err := o.tasks.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get project info
	projects, err := o.github.GetProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}

	var owner, repo string
	for _, p := range projects {
		if p.ID == task.ProjectID {
			owner = p.GitHubOwner
			repo = p.GitHubRepo
			break
		}
	}

	if owner == "" {
		return fmt.Errorf("project not found for task")
	}

	// First, pull main into the worktree to check for conflicts
	if err := o.repos.PullMainIntoWorktree(owner, repo, taskID); err != nil {
		// Check if it's a merge conflict - return it for the handler to show UI
		if conflictErr, ok := err.(*git.ErrMergeConflict); ok {
			slog.Info("[ORCHESTRATOR] Merge conflict detected",
				"task_id", taskID, "files", conflictErr.ConflictedFiles)
			return conflictErr
		}
		return fmt.Errorf("failed to pull main: %w", err)
	}

	// No conflicts - commit the merge and push
	if err := o.repos.CommitAndPush(taskID, "Merge main into branch"); err != nil {
		slog.Warn("[ORCHESTRATOR] No changes to commit after merge", "task_id", taskID)
	}

	// Now merge to main
	branchName, err := o.repos.MergeToMain(owner, repo, taskID)
	if err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	slog.Info("[ORCHESTRATOR] Merged task to main", "task_id", taskID, "branch", branchName)

	// Update task status to done
	if err := o.tasks.UpdateStatus(ctx, taskID, models.StatusDone); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Clean up the worktree
	if err := o.repos.RemoveWorktree(taskID); err != nil {
		slog.Warn("[ORCHESTRATOR] Failed to remove worktree after merge", "task_id", taskID, "error", err)
	}

	// Emit status change event
	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   "status_change",
	})

	return nil
}

// GetConflictDetails returns the content of conflicted files for UI display.
func (o *Orchestrator) GetConflictDetails(ctx context.Context, taskID string, files []string) ([]ConflictFile, error) {
	worktreePath := o.repos.WorktreePath(taskID)
	result := make([]ConflictFile, 0, len(files))

	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(worktreePath, file))
		if err != nil {
			continue
		}

		// Parse the conflict markers
		cf := parseConflictFile(file, string(content))
		result = append(result, cf)
	}

	return result, nil
}

// ConflictFile represents a file with merge conflicts.
type ConflictFile struct {
	Path     string
	Ours     string // Current branch (HEAD)
	Theirs   string // Incoming (origin/main)
	Original string // Full content with markers
}

// parseConflictFile parses conflict markers from file content.
func parseConflictFile(path, content string) ConflictFile {
	cf := ConflictFile{
		Path:     path,
		Original: content,
	}

	// Simple parsing - extract ours and theirs sections
	lines := strings.Split(content, "\n")
	var ours, theirs []string
	inOurs, inTheirs := false, false

	for _, line := range lines {
		if strings.HasPrefix(line, "<<<<<<<") {
			inOurs = true
			continue
		}
		if strings.HasPrefix(line, "|||||||") {
			inOurs = false
			continue
		}
		if line == "=======" {
			inOurs = false
			inTheirs = true
			continue
		}
		if strings.HasPrefix(line, ">>>>>>>") {
			inTheirs = false
			continue
		}

		if inOurs {
			ours = append(ours, line)
		} else if inTheirs {
			theirs = append(theirs, line)
		}
	}

	cf.Ours = strings.Join(ours, "\n")
	cf.Theirs = strings.Join(theirs, "\n")
	return cf
}

// ResolveConflict resolves a single file conflict with the chosen version.
func (o *Orchestrator) ResolveConflict(ctx context.Context, taskID, filePath, resolution string) error {
	worktreePath := o.repos.WorktreePath(taskID)
	fullPath := filepath.Join(worktreePath, filePath)

	// Write the resolved content
	if err := os.WriteFile(fullPath, []byte(resolution), 0644); err != nil {
		return fmt.Errorf("failed to write resolved file: %w", err)
	}

	// Stage the file
	cmd := exec.Command("git", "add", filePath)
	cmd.Dir = worktreePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stage resolved file: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[ORCHESTRATOR] Resolved conflict", "task_id", taskID, "file", filePath)
	return nil
}

// AbortMerge aborts the current merge operation.
func (o *Orchestrator) AbortMerge(ctx context.Context, taskID string) error {
	return o.repos.AbortMerge(taskID)
}

// CompleteMergeResolution finishes the merge after all conflicts are resolved.
func (o *Orchestrator) CompleteMergeResolution(ctx context.Context, taskID string) error {
	// Get task info
	task, err := o.tasks.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get project info
	projects, err := o.github.GetProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to get projects: %w", err)
	}

	var owner, repo string
	for _, p := range projects {
		if p.ID == task.ProjectID {
			owner = p.GitHubOwner
			repo = p.GitHubRepo
			break
		}
	}

	// Commit the merge resolution
	if err := o.repos.CommitMergeResolution(taskID, "Resolve merge conflicts"); err != nil {
		return fmt.Errorf("failed to commit resolution: %w", err)
	}

	// Now merge to main
	branchName, err := o.repos.MergeToMain(owner, repo, taskID)
	if err != nil {
		return fmt.Errorf("failed to merge to main: %w", err)
	}

	slog.Info("[ORCHESTRATOR] Merged task to main after conflict resolution", "task_id", taskID, "branch", branchName)

	// Update task status to done
	if err := o.tasks.UpdateStatus(ctx, taskID, models.StatusDone); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Clean up the worktree
	if err := o.repos.RemoveWorktree(taskID); err != nil {
		slog.Warn("[ORCHESTRATOR] Failed to remove worktree", "task_id", taskID, "error", err)
	}

	// Emit status change event
	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   "status_change",
	})

	return nil
}

// CreatePR creates a GitHub Pull Request for a task and marks it done.
func (o *Orchestrator) CreatePR(ctx context.Context, taskID string) (string, error) {
	// Get task info
	task, err := o.tasks.Get(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("task not found: %w", err)
	}

	// Get project info
	projects, err := o.github.GetProjects(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	var owner, repo string
	for _, p := range projects {
		if p.ID == task.ProjectID {
			owner = p.GitHubOwner
			repo = p.GitHubRepo
			break
		}
	}

	if owner == "" {
		return "", fmt.Errorf("project not found for task")
	}

	// Get the branch name from the worktree
	branchName, err := o.repos.GetCurrentBranch(taskID)
	if err != nil {
		return "", fmt.Errorf("failed to get branch name: %w", err)
	}
	branchName = strings.TrimSpace(branchName)

	// Create the PR - use Title for both title and body
	prURL, err := o.github.CreatePullRequest(ctx, owner, repo, branchName, task.Title, task.Intent)
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %w", err)
	}

	slog.Info("[ORCHESTRATOR] Created PR", "task_id", taskID, "pr_url", prURL)

	// Update task status to done
	if err := o.tasks.UpdateStatus(ctx, taskID, models.StatusDone); err != nil {
		return prURL, fmt.Errorf("PR created but failed to update task status: %w", err)
	}

	// Emit status change event
	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   "status_change",
	})

	return prURL, nil
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

// CleanupTask removes the worktree for a task.
func (o *Orchestrator) CleanupTask(taskID string) error {
	return o.repos.RemoveWorktree(taskID)
}

// emit sends an event to the UI and stores it in DB.
func (o *Orchestrator) emit(taskID, level, message string) {
	// Store log in DB for persistence
	ctx := context.Background()
	if err := o.tasks.AddLog(ctx, taskID, level, message); err != nil {
		slog.Error("Failed to store log", "task_id", taskID, "error", err)
	}

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

// appendUserMessage appends a user message to the existing message history JSON
func appendUserMessage(historyJSON, message string) string {
	// Parse existing history
	var messages []map[string]any
	if historyJSON != "" {
		if err := json.Unmarshal([]byte(historyJSON), &messages); err != nil {
			messages = []map[string]any{}
		}
	}

	// Append user message
	userMsg := map[string]any{
		"role": "user",
		"content": []map[string]any{
			{"type": "text", "text": message},
		},
	}
	messages = append(messages, userMsg)

	// Marshal back to JSON
	result, err := json.Marshal(messages)
	if err != nil {
		return historyJSON
	}
	return string(result)
}
