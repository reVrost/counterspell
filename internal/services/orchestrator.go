// Package services package contains all the services used by the server.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lithammer/shortuuid/v4"
	"github.com/panjf2000/ants/v2"
	"github.com/revrost/code/counterspell/internal/agent"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
	"github.com/revrost/code/counterspell/internal/llm"
	"github.com/revrost/code/counterspell/internal/models"
)

// TaskResult represents the result of a completed task.
type TaskResult struct {
	TaskID         string
	UserID         string
	RunID          string
	Success        bool
	Output         string
	GitDiff        string
	MessageHistory string
	ArtifactPath   string
	Error          string
}

// TaskJob represents a job submitted to the worker pool.
type TaskJob struct {
	TaskID    string
	UserID    string
	ProjectID string
	Intent    string
	ModelID   string
	Owner     string
	Repo      string
	Token     string
	RunID     string // Current agent run ID
	ResultCh  chan<- TaskResult
}

// Orchestrator manages task execution with agents.
type Orchestrator struct {
	db         *db.DB
	tasks      *TaskService
	github     *GitHubService
	repos      *RepoManager
	events     *EventBus
	settings   *SettingsService
	dataDir    string
	userID     string
	workerPool *ants.Pool
	resultCh   chan TaskResult
	running    map[string]context.CancelFunc
	mu         sync.Mutex
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(
	database *db.DB,
	tasks *TaskService,
	github *GitHubService,
	events *EventBus,
	settings *SettingsService,
	dataDir string,
	userID string,
) (*Orchestrator, error) {
	slog.Info("[ORCHESTRATOR] Creating new orchestrator", "data_dir", dataDir, "user_id", userID)
	// Create worker pool with 5 workers
	pool, err := ants.NewPool(5, ants.WithPreAlloc(false))
	if err != nil {
		return nil, fmt.Errorf("failed to create worker pool: %w", err)
	}

	orch := &Orchestrator{
		db:         database,
		tasks:      tasks,
		github:     github,
		events:     events,
		settings:   settings,
		repos:      NewRepoManager(dataDir),
		dataDir:    dataDir,
		userID:     userID,
		workerPool: pool,
		resultCh:   make(chan TaskResult, 100),
		running:    make(map[string]context.CancelFunc),
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
	projects, err := o.github.GetProjects(ctx, o.userID)
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
		return nil, fmt.Errorf("project not found (ID: %s), please select a valid project", projectID)
	}

	// Get GitHub token
	conn, err := o.github.GetActiveConnection(ctx, o.userID)
	if err != nil {
		return nil, fmt.Errorf("no GitHub connection: %w", err)
	}
	if conn == nil {
		return nil, fmt.Errorf("no GitHub connection found")
	}
	tokenLen := len(conn.Token)
	slog.Info("GitHub token retrieved", "token_length", tokenLen, "has_token", tokenLen > 0)
	if tokenLen == 0 {
		slog.Error("GitHub token is empty - need to configure in settings")
		return nil, fmt.Errorf("GitHub token is not configured, please add a GitHub connection in settings")
	}

	// Create task in DB (after validating project exists)
	task, err := o.tasks.Create(ctx, o.userID, projectID, intent, intent)
	if err != nil {
		slog.Error("Failed to create task in database", "project_id", projectID, "error", err)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	slog.Info("Task created successfully", "task_id", task.ID, "project_id", projectID, "owner", owner, "repo", repo, "model_id", modelID)

	// Emit task_created event to trigger feed refresh
	o.events.Publish(models.Event{
		TaskID: task.ID,
		Type:   models.EventTypeTaskCreated,
	})

	slog.Info("Submitting task to worker pool", "task_id", task.ID, "intent", intent)

	// Submit job to worker pool
	job := &TaskJob{
		TaskID:    task.ID,
		UserID:    o.userID,
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

	// Create agent run record
	runID := shortuuid.New()
	job.RunID = runID
	now := time.Now()
	if err := o.createAgentRun(ctx, runID, job.TaskID, "execution", job.Intent, now); err != nil {
		slog.Error("[AGENT LOOP] Failed to create agent run", "task_id", job.TaskID, "error", err)
	}

	// Update task status to in_progress
	slog.Info("[AGENT LOOP] Updating task status to in_progress", "task_id", job.TaskID)
	if err := o.tasks.UpdateStatusAndStep(ctx, job.UserID, job.TaskID, models.StatusInProgress, "execution"); err != nil {
		slog.Error("[AGENT LOOP] Failed to update task status", "task_id", job.TaskID, "error", err)
		o.emitError(job.TaskID, "Failed to update status")
		o.failAgentRun(ctx, runID, "Failed to update status", "", now)
		job.ResultCh <- TaskResult{
			TaskID:  job.TaskID,
			UserID:  job.UserID,
			RunID:   runID,
			Success: false,
			Error:   "Failed to update status",
		}
		return
	}

	// Update run status to running
	_ = o.updateAgentRunStatus(ctx, runID, models.RunStatusRunning, now)

	o.emit(job.TaskID, "plan", "Starting task execution...")
	o.emit(job.TaskID, "info", fmt.Sprintf("Preparing %s/%s...", job.Owner, job.Repo))

	// Clone/fetch repo
	slog.Info("[AGENT LOOP] Ensuring repo exists", "task_id", job.TaskID, "owner", job.Owner, "repo", job.Repo)
	repoPath, err := o.repos.EnsureRepo(job.Owner, job.Repo, job.Token)
	if err != nil {
		slog.Error("[AGENT LOOP] Failed to prepare repo", "task_id", job.TaskID, "error", err)
		o.emitError(job.TaskID, "Failed to prepare repo: "+err.Error())
		o.failAgentRun(ctx, runID, "Failed to prepare repo: "+err.Error(), "", time.Now())
		job.ResultCh <- TaskResult{
			TaskID:  job.TaskID,
			UserID:  job.UserID,
			RunID:   runID,
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
		o.failAgentRun(ctx, runID, "Failed to create worktree: "+err.Error(), "", time.Now())
		job.ResultCh <- TaskResult{
			TaskID:  job.TaskID,
			UserID:  job.UserID,
			RunID:   runID,
			Success: false,
			Error:   "Failed to create worktree: " + err.Error(),
		}
		return
	}
	slog.Info("[AGENT LOOP] Worktree created successfully", "task_id", job.TaskID, "worktree_path", worktreePath)

	o.emit(job.TaskID, "success", fmt.Sprintf("Workspace ready: %s", branchName))

	// Get API key and create provider
	slog.Info("[AGENT LOOP] Getting settings", "task_id", job.TaskID)
	settings, err := o.settings.GetSettings(ctx, job.UserID)
	if err != nil {
		slog.Error("[AGENT LOOP] Failed to get settings", "task_id", job.TaskID, "error", err)
		o.emitError(job.TaskID, "Failed to get settings")
		o.failAgentRun(ctx, runID, "Failed to get settings", "", time.Now())
		job.ResultCh <- TaskResult{
			TaskID:  job.TaskID,
			UserID:  job.UserID,
			RunID:   runID,
			Success: false,
			Error:   "Failed to get settings",
		}
		return
	}
	slog.Info("[AGENT LOOP] Got settings", "task_id", job.TaskID,
		"openrouter_key_len", len(settings.OpenRouterKey),
		"zai_key_len", len(settings.ZaiKey),
		"anthropic_key_len", len(settings.AnthropicKey),
		"openai_key_len", len(settings.OpenAIKey),
		"agent_backend", settings.GetAgentBackend())

	// Create agent backend based on settings
	agentBackend := settings.GetAgentBackend()
	slog.Info("[AGENT LOOP] Creating agent backend", "task_id", job.TaskID, "backend", agentBackend)

	var backend agent.Backend
	var agentOutput, messageHistory string

	callback := func(event agent.StreamEvent) {
		slog.Debug("[AGENT LOOP] Agent event", "task_id", job.TaskID, "event_type", event.Type)
		o.handleAgentEvent(job.TaskID, event)
	}

	// Claude Code integration configuration
	type claudeCodeProvider struct {
		baseURL   string
		getKey    func(*models.UserSettings) string
		supported bool
	}

	claudeCodeProviders := map[string]claudeCodeProvider{
		"anthropic":  {getKey: func(s *models.UserSettings) string { return s.AnthropicKey }, supported: true},
		"zai":        {baseURL: "https://api.z.ai/api/anthropic", getKey: func(s *models.UserSettings) string { return s.ZaiKey }, supported: true},
		"z":          {baseURL: "https://api.z.ai/api/anthropic", getKey: func(s *models.UserSettings) string { return s.ZaiKey }, supported: true},
		"openrouter": {baseURL: "https://openrouter.ai/api", getKey: func(s *models.UserSettings) string { return s.OpenRouterKey }, supported: true},
		"o":          {baseURL: "https://openrouter.ai/api", getKey: func(s *models.UserSettings) string { return s.OpenRouterKey }, supported: true},
	}

	// Check if claude-code backend should be used
	providerPrefix, modelName := llm.ParseModelID(job.ModelID)
	providerConfig, isSupported := claudeCodeProviders[providerPrefix]
	useClaudeCode := agentBackend == models.AgentBackendClaudeCode && isSupported

	if agentBackend == models.AgentBackendClaudeCode && !useClaudeCode {
		slog.Info("[AGENT LOOP] Claude Code requested but unsupported provider, falling back to native",
			"task_id", job.TaskID, "provider", providerPrefix, "model", modelName)
		o.emit(job.TaskID, "plan", "Using native backend (Claude Code only supports Anthropic/Z.AI/OpenRouter)")
	}

	switch {
	case useClaudeCode:
		// Use Claude Code CLI backend
		o.emit(job.TaskID, "plan", "Starting Claude Code agent...")

		// Configure based on provider
		opts := []agent.ClaudeCodeOption{
			agent.WithClaudeWorkDir(worktreePath),
			agent.WithClaudeCallback(callback),
		}

		// Apply provider configuration
		if providerConfig.baseURL != "" {
			opts = append(opts, agent.WithBaseURL(providerConfig.baseURL))
		}
		opts = append(opts, agent.WithAPIKey(providerConfig.getKey(settings)))
		if modelName != "" {
			opts = append(opts, agent.WithModel(modelName))
		}

		slog.Info("[AGENT LOOP] Using Claude Code", "task_id", job.TaskID, "provider", providerPrefix, "model", modelName)

		claudeBackend, err := agent.NewClaudeCodeBackend(opts...)
		if err != nil {
			slog.Error("[AGENT LOOP] Failed to create Claude Code backend", "task_id", job.TaskID, "error", err)
			o.emitError(job.TaskID, "Claude Code not available: "+err.Error())
			o.failAgentRun(ctx, runID, "Claude Code not available: "+err.Error(), "", time.Now())
			job.ResultCh <- TaskResult{
				TaskID:  job.TaskID,
				UserID:  job.UserID,
				RunID:   runID,
				Success: false,
				Error:   "Claude Code not available: " + err.Error(),
			}
			return
		}
		backend = claudeBackend
		defer func() { _ = claudeBackend.Close() }()

	default:
		// Use native backend (Counterspell)
		slog.Info("[AGENT LOOP] Using native backend", "task_id", job.TaskID, "provider", providerPrefix, "model", modelName)

		var provider llm.Provider
		switch providerPrefix {
		case "openrouter", "o":
			if settings.OpenRouterKey == "" {
				o.emitError(job.TaskID, "OpenRouter API key not configured")
				o.failAgentRun(ctx, runID, "OpenRouter API key not configured", "", time.Now())
				job.ResultCh <- TaskResult{
					TaskID:  job.TaskID,
					UserID:  job.UserID,
					RunID:   runID,
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
				o.failAgentRun(ctx, runID, "Z.ai API key not configured", "", time.Now())
				job.ResultCh <- TaskResult{
					TaskID:  job.TaskID,
					UserID:  job.UserID,
					RunID:   runID,
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
				o.failAgentRun(ctx, runID, "Anthropic API key not configured", "", time.Now())
				job.ResultCh <- TaskResult{
					TaskID:  job.TaskID,
					UserID:  job.UserID,
					RunID:   runID,
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
				o.failAgentRun(ctx, runID, "No API key configured", "", time.Now())
				job.ResultCh <- TaskResult{
					TaskID:  job.TaskID,
					UserID:  job.UserID,
					RunID:   runID,
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

		o.emit(job.TaskID, "plan", fmt.Sprintf("Starting Counterspell agent with model: %s...", finalModel))

		nativeBackend, err := agent.NewNativeBackend(
			agent.WithProvider(provider),
			agent.WithWorkDir(worktreePath),
			agent.WithCallback(callback),
		)
		if err != nil {
			slog.Error("[AGENT LOOP] Failed to create native backend", "task_id", job.TaskID, "error", err)
			o.emitError(job.TaskID, "Failed to create agent: "+err.Error())
			o.failAgentRun(ctx, runID, "Failed to create agent: "+err.Error(), "", time.Now())
			job.ResultCh <- TaskResult{
				TaskID:  job.TaskID,
				UserID:  job.UserID,
				RunID:   runID,
				Success: false,
				Error:   "Failed to create agent: " + err.Error(),
			}
			return
		}
		backend = nativeBackend
		defer func() { _ = nativeBackend.Close() }()
	}

	// Run the agent
	slog.Info("[AGENT LOOP] Starting agent execution", "task_id", job.TaskID, "backend", agentBackend)
	runErr := backend.Run(ctx, job.Intent)
	if runErr != nil {
		slog.Error("[AGENT LOOP] Agent execution failed", "task_id", job.TaskID, "error", runErr)
		if ctx.Err() != nil {
			o.emit(job.TaskID, "info", "Task cancelled")
		} else {
			o.emitError(job.TaskID, "Agent failed: "+runErr.Error())
		}
		o.failAgentRun(ctx, runID, "Agent failed: "+runErr.Error(), backend.GetState(), time.Now())
		job.ResultCh <- TaskResult{
			TaskID:  job.TaskID,
			UserID:  job.UserID,
			RunID:   runID,
			Success: false,
			Error:   "Agent failed: " + runErr.Error(),
		}
		return
	}
	slog.Info("[AGENT LOOP] Agent execution completed successfully", "task_id", job.TaskID)

	// Extract results from backend
	agentOutput = backend.FinalMessage()
	messageHistory = backend.GetState()

	// Commit changes (push happens on merge/PR action)
	o.emit(job.TaskID, "info", "Committing changes...")

	commitMsg := fmt.Sprintf("feat: %s\n\nTask ID: %s", job.Intent, job.TaskID)
	if err := o.repos.Commit(job.TaskID, commitMsg); err != nil {
		o.emit(job.TaskID, "info", "No changes to commit: "+err.Error())
	} else {
		o.emit(job.TaskID, "success", fmt.Sprintf("Changes committed to branch: %s", branchName))
	}

	// Get git diff
	gitDiff, err := o.repos.GetDiff(job.TaskID)
	if err != nil {
		slog.Warn("Failed to get git diff", "task_id", job.TaskID, "error", err)
		gitDiff = ""
	}

	// Send result to channel for processing
	job.ResultCh <- TaskResult{
		TaskID:         job.TaskID,
		UserID:         job.UserID,
		RunID:          runID,
		Success:        true,
		Output:         agentOutput,
		GitDiff:        gitDiff,
		MessageHistory: messageHistory,
	}

	slog.Info("Task completed", "task_id", job.TaskID, "repo", fmt.Sprintf("%s/%s", job.Owner, job.Repo), "path", repoPath)
}

// processResults processes task results from the channel.
func (o *Orchestrator) processResults() {
	slog.Info("[ORCHESTRATOR] processResults goroutine started")
	for result := range o.resultCh {
		slog.Info("[ORCHESTRATOR] Received result from worker", "task_id", result.TaskID, "success", result.Success)
		ctx := context.Background()

		// Update task status - after execution, moves to review
		status := models.StatusReview
		if !result.Success {
			status = models.StatusFailed
		}

		if err := o.tasks.UpdateStatus(ctx, result.UserID, result.TaskID, status); err != nil {
			slog.Error("Failed to update task status", "task_id", result.TaskID, "error", err)
			continue
		}

		// Complete the agent run
		if result.RunID != "" {
			if result.Success {
				_ = o.completeAgentRun(ctx, result.RunID, result.Output, result.MessageHistory, "", time.Now())
			}
		}

		if result.Success {
			o.emit(result.TaskID, "success", "Task complete - ready for review")
		} else {
			o.emitError(result.TaskID, result.Error)
		}

		// Emit status_change event to trigger SSE updates
		o.events.Publish(models.Event{
			TaskID: result.TaskID,
			Type:   models.EventTypeStatusChange,
		})

		slog.Info("[ORCHESTRATOR] Result processed", "task_id", result.TaskID)
	}
	slog.Info("[ORCHESTRATOR] processResults goroutine ended")
}

// handleAgentEvent converts agent events to UI events.
func (o *Orchestrator) handleAgentEvent(taskID string, event agent.StreamEvent) {
	slog.Info("[ORCHESTRATOR] handleAgentEvent", "task_id", taskID, "event_type", event.Type, "content_len", len(event.Content))
	switch event.Type {
	case agent.EventPlan:
		o.emit(taskID, "plan", event.Content)
	case agent.EventTool:
		o.emit(taskID, "code", fmt.Sprintf("[%s] %s", event.Tool, truncate(event.Args, 60)))
	case agent.EventResult:
		o.emit(taskID, "info", truncate(event.Content, 100))
	case agent.EventText:
		o.emit(taskID, "info", event.Content)
	case agent.EventError:
		o.emitError(taskID, event.Content)
	case agent.EventDone:
		o.emit(taskID, "success", event.Content)
	case agent.EventMessages:
		// Publish message history for live agent panel updates
		o.events.Publish(models.Event{
			TaskID: taskID,
			Type:   models.EventTypeAgentUpdate,
			Data:   event.Messages, // JSON message history
		})
	case agent.EventTodo:
		// Publish todo list for live todo panel updates
		o.events.Publish(models.Event{
			TaskID: taskID,
			Type:   models.EventTypeTodo,
			Data:   event.Content, // JSON todo list
		})
	}
}

// ContinueTask continues an existing task with a follow-up message.
func (o *Orchestrator) ContinueTask(ctx context.Context, taskID, followUpMessage, modelID string) error {
	// Get existing task
	task, err := o.tasks.Get(ctx, o.userID, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get the latest agent run to retrieve message history
	latestRun, err := o.getLatestAgentRun(ctx, taskID, "execution")
	if err != nil {
		slog.Warn("[CHAT] No previous run found", "task_id", taskID, "error", err)
	}

	var messageHistory string
	if latestRun != nil {
		messageHistory = string(latestRun.MessageHistory)
	}

	// Get project info
	projects, err := o.github.GetProjects(ctx, o.userID)
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
	conn, err := o.github.GetActiveConnection(ctx, o.userID)
	if err != nil {
		return fmt.Errorf("no GitHub connection: %w", err)
	}
	if conn == nil {
		return fmt.Errorf("no GitHub connection found")
	}

	// Update task status back to in_progress
	if err := o.tasks.UpdateStatus(ctx, o.userID, taskID, models.StatusInProgress); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Immediately emit agent_update with user message appended so it shows right away
	updatedHistory := appendUserMessage(messageHistory, followUpMessage)
	slog.Info("[CHAT] Publishing agent_update with user message", "task_id", taskID, "history_len", len(updatedHistory))
	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   models.EventTypeAgentUpdate,
		Data:   updatedHistory,
	})

	// Emit task_created event to trigger feed refresh (after agent_update is cached)
	slog.Info("[CHAT] Publishing task_created", "task_id", taskID)
	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   models.EventTypeTaskCreated,
	})

	slog.Info("[CHAT] Events published, submitting to worker pool", "task_id", taskID, "follow_up", followUpMessage)

	// Submit job to worker pool - for now, treat as new execution
	job := &TaskJob{
		TaskID:    taskID,
		UserID:    o.userID,
		ProjectID: task.ProjectID,
		Intent:    followUpMessage,
		ModelID:   modelID,
		Owner:     owner,
		Repo:      repo,
		Token:     conn.Token,
		ResultCh:  o.resultCh,
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
func (o *Orchestrator) MergeTask(ctx context.Context, taskID string) error {
	// Get task to find project info
	task, err := o.tasks.Get(ctx, o.userID, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get project info
	projects, err := o.github.GetProjects(ctx, o.userID)
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
		if conflictErr, ok := err.(*ErrMergeConflict); ok {
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
	if err := o.tasks.UpdateStatus(ctx, o.userID, taskID, models.StatusDone); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Clean up the worktree
	if err := o.repos.RemoveWorktree(taskID); err != nil {
		slog.Warn("[ORCHESTRATOR] Failed to remove worktree after merge", "task_id", taskID, "error", err)
	}

	// Emit status change event
	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   models.EventTypeStatusChange,
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
	task, err := o.tasks.Get(ctx, o.userID, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get project info
	projects, err := o.github.GetProjects(ctx, o.userID)
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
	if err := o.tasks.UpdateStatus(ctx, o.userID, taskID, models.StatusDone); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Clean up the worktree
	if err := o.repos.RemoveWorktree(taskID); err != nil {
		slog.Warn("[ORCHESTRATOR] Failed to remove worktree", "task_id", taskID, "error", err)
	}

	// Emit status change event
	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   models.EventTypeStatusChange,
	})

	return nil
}

// CreatePR creates a GitHub Pull Request for a task and marks it done.
func (o *Orchestrator) CreatePR(ctx context.Context, taskID string) (string, error) {
	// Get task info
	task, err := o.tasks.Get(ctx, o.userID, taskID)
	if err != nil {
		return "", fmt.Errorf("task not found: %w", err)
	}

	// Get project info
	projects, err := o.github.GetProjects(ctx, o.userID)
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

	// Push the branch to remote before creating PR
	if err := o.repos.PushBranch(taskID); err != nil {
		return "", fmt.Errorf("failed to push branch: %w", err)
	}

	// Create the PR - use Title for both title and body
	prURL, err := o.github.CreatePullRequest(ctx, o.userID, owner, repo, branchName, task.Title, task.Intent)
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %w", err)
	}

	slog.Info("[ORCHESTRATOR] Created PR", "task_id", taskID, "pr_url", prURL)

	// Update task status to done
	if err := o.tasks.UpdateStatus(ctx, o.userID, taskID, models.StatusDone); err != nil {
		return prURL, fmt.Errorf("PR created but failed to update task status: %w", err)
	}

	// Emit status change event
	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   models.EventTypeStatusChange,
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

// emit sends an event to the UI.
func (o *Orchestrator) emit(taskID, level, message string) {
	payload := map[string]string{
		"level":   level,
		"message": message,
	}
	data, _ := json.Marshal(payload)

	o.events.Publish(models.Event{
		TaskID: taskID,
		Type:   models.EventTypeLog,
		Data:   string(data),
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

// SearchProjectFiles searches for files in a project using fuzzy matching.
func (o *Orchestrator) SearchProjectFiles(ctx context.Context, projectID, query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}

	// Get project info
	projects, err := o.github.GetProjects(ctx, o.userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	var owner, repo string
	for _, p := range projects {
		if p.ID == projectID {
			owner = p.GitHubOwner
			repo = p.GitHubRepo
			break
		}
	}

	if owner == "" {
		return nil, fmt.Errorf("project not found")
	}

	repoPath := o.repos.RepoPath(owner, repo)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository not cloned yet")
	}

	// Collect all file paths
	var files []string
	err = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		// Skip hidden directories and common non-source dirs
		name := info.Name()
		if info.IsDir() {
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			return nil
		}
		// Get relative path from repo root
		relPath, _ := filepath.Rel(repoPath, path)
		files = append(files, relPath)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk repo: %w", err)
	}

	// If no query, return first N files sorted alphabetically
	if query == "" {
		if len(files) > limit {
			files = files[:limit]
		}
		return files, nil
	}

	// Fuzzy search files
	matches := fuzzySearch(files, query, limit)
	return matches, nil
}

// fuzzySearch performs fuzzy matching on file paths and returns top N matches.
func fuzzySearch(files []string, query string, limit int) []string {
	type scored struct {
		path  string
		score int
	}

	var results []scored
	queryLower := strings.ToLower(query)

	for _, f := range files {
		fLower := strings.ToLower(f)
		score := fuzzyScore(fLower, queryLower)
		if score > 0 {
			results = append(results, scored{path: f, score: score})
		}
	}

	// Sort by score descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Return top N
	var out []string
	for i := 0; i < len(results) && i < limit; i++ {
		out = append(out, results[i].path)
	}
	return out
}

// fuzzyScore computes a simple fuzzy match score.
func fuzzyScore(text, pattern string) int {
	if len(pattern) == 0 {
		return 1
	}
	if len(text) == 0 {
		return 0
	}

	score := 0
	patternIdx := 0
	prevMatchIdx := -1
	consecutiveBonus := 0

	for i := 0; i < len(text) && patternIdx < len(pattern); i++ {
		if text[i] == pattern[patternIdx] {
			score += 1

			// Bonus for consecutive matches
			if prevMatchIdx == i-1 {
				consecutiveBonus++
				score += consecutiveBonus
			} else {
				consecutiveBonus = 0
			}

			// Bonus for matching at start or after separator
			if i == 0 || text[i-1] == '/' || text[i-1] == '_' || text[i-1] == '-' || text[i-1] == '.' {
				score += 2
			}

			prevMatchIdx = i
			patternIdx++
		}
	}

	// All pattern characters must match
	if patternIdx < len(pattern) {
		return 0
	}

	return score
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

// Agent run helper methods

func (o *Orchestrator) createAgentRun(ctx context.Context, id, taskID, step, input string, now time.Time) error {
	_, err := o.db.Queries.CreateAgentRun(ctx, sqlc.CreateAgentRunParams{
		ID:        id,
		TaskID:    taskID,
		Step:      step,
		AgentID:   pgtype.Text{},
		Status:    string(models.RunStatusPending),
		Input:     pgtype.Text{String: input, Valid: input != ""},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	return err
}

func (o *Orchestrator) updateAgentRunStatus(ctx context.Context, id string, status models.AgentRunStatus, startedAt time.Time) error {
	return o.db.Queries.UpdateAgentRunStatus(ctx, sqlc.UpdateAgentRunStatusParams{
		Status:    string(status),
		StartedAt: pgtype.Timestamptz{Time: startedAt, Valid: true},
		ID:        id,
	})
}

func (o *Orchestrator) completeAgentRun(ctx context.Context, id, output, messageHistory, artifactPath string, completedAt time.Time) error {
	return o.db.Queries.CompleteAgentRun(ctx, sqlc.CompleteAgentRunParams{
		Output:         pgtype.Text{String: output, Valid: output != ""},
		MessageHistory: []byte(messageHistory),
		ArtifactPath:   pgtype.Text{String: artifactPath, Valid: artifactPath != ""},
		CompletedAt:    pgtype.Timestamptz{Time: completedAt, Valid: true},
		ID:             id,
	})
}

func (o *Orchestrator) failAgentRun(ctx context.Context, id, errMsg, messageHistory string, completedAt time.Time) error {
	return o.db.Queries.FailAgentRun(ctx, sqlc.FailAgentRunParams{
		Error:          pgtype.Text{String: errMsg, Valid: errMsg != ""},
		MessageHistory: []byte(messageHistory),
		CompletedAt:    pgtype.Timestamptz{Time: completedAt, Valid: true},
		ID:             id,
	})
}

func (o *Orchestrator) getLatestAgentRun(ctx context.Context, taskID, step string) (*sqlc.AgentRun, error) {
	run, err := o.db.Queries.GetLatestRunForStep(ctx, sqlc.GetLatestRunForStepParams{
		TaskID: taskID,
		Step:   step,
	})
	if err != nil {
		return nil, err
	}
	return &run, nil
}
