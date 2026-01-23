package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/revrost/code/counterspell/internal/agent"
	"github.com/revrost/code/counterspell/internal/llm"
	"github.com/revrost/code/counterspell/internal/models"
)

// ConflictFile represents a merge conflict.
type ConflictFile struct {
	Path     string `json:"path"`
	Current  string `json:"current"`
	Incoming string `json:"incoming"`
	Base     string `json:"base"`
}

// TaskResult represents result of a completed task.
type TaskResult struct {
	TaskID      string
	Success     bool
	AgentOutput string
	GitDiff     string
	Error       string
}

// TaskJob represents a job submitted to the worker pool.
type TaskJob struct {
	TaskID         string
	ProjectID      string
	Intent         string
	ModelID        string // Format: "provider:model" e.g., "anthropic:claude-opus-4-5"
	Owner          string
	Repo           string
	Token          string
	MessageHistory string // Only for continuations
	IsContinuation bool
	ResultCh       chan<- TaskResult
}

// Orchestrator manages task execution with agents.
type Orchestrator struct {
	repo            *Repository
	gitReposManager *GitManager
	eventBus        *EventBus
	settings        *SettingsService
	github          *GitHubService
	dataDir         string
	workerPool      *ants.Pool
	resultCh        chan TaskResult
	running         map[string]context.CancelFunc
	mu              sync.Mutex
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(
	repo *Repository,
	eventBus *EventBus,
	settings *SettingsService,
	github *GitHubService,
	dataDir string,
) (*Orchestrator, error) {
	userID := "default" // Hardcoded for local-first single-tenant mode
	slog.Info("[ORCHESTRATOR] Creating new orchestrator", "data_dir", dataDir, "user_id", userID)
	// Create worker pool with 5 workers
	pool, err := ants.NewPool(5, ants.WithPreAlloc(false))
	if err != nil {
		return nil, err
	}

	orch := &Orchestrator{
		repo:     repo,
		eventBus: eventBus,
		settings: settings,
		github:   github,

		gitReposManager: NewGitManager(dataDir),
		dataDir:         dataDir,

		// Worker related fields
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

// Shutdown gracefully shuts down orchestrator.
func (o *Orchestrator) Shutdown() {
	slog.Info("[ORCHESTRATOR] Shutting down")

	// Cancel all running tasks
	o.mu.Lock()
	for taskID, cancel := range o.running {
		slog.Info("[ORCHESTRATOR] Cancelling running task", "task_id", taskID)
		cancel()
	}
	o.running = make(map[string]context.CancelFunc)
	o.mu.Unlock()

	// Release worker pool (prevents new tasks from starting)
	o.workerPool.Release()

	// Close result channel to unblock processResults goroutine
	close(o.resultCh)

	slog.Info("[ORCHESTRATOR] Shutdown complete")
}

// StartTask creates a task and begins execution.
func (o *Orchestrator) StartTask(ctx context.Context, projectID, intent, modelID string) (string, error) {
	// 1. Resolve projectID to a repository and ensure it's cloned
	var token string
	var owner, repoName string
	if projectID == "" {
		return "", fmt.Errorf("project_id is required")
	}
	// Look up repo in DB
	repo, err := o.repo.GetRepository(ctx, projectID)
	if err == nil {
		// Get connection for token
		conn, err := o.repo.GetGithubConnectionByID(ctx, repo.ConnectionID)
		if err == nil {
			token = conn.AccessToken
			owner = repo.Owner
			repoName = repo.Name
			slog.Info("[ORCHESTRATOR] Found repository and connection", "repo", repo.FullName, "owner", owner)

			// Ensure repo exists
			_, err = o.gitReposManager.EnsureRepo(owner, repoName, token)
			if err != nil {
				return "", fmt.Errorf("failed to ensure repo: %w", err)
			}
		}
	}

	// Create task in database
	task, err := o.repo.Create(ctx, projectID, intent)
	if err != nil {
		return "", err
	}
	taskID := task.ID

	slog.Info("[ORCHESTRATOR] Task created", "task_id", taskID, "project_id", projectID, "intent", intent)

	// Submit job to worker pool
	job := TaskJob{
		TaskID:         taskID,
		ProjectID:      projectID,
		Intent:         intent,
		ModelID:        modelID,
		Owner:          owner,
		Repo:           repoName,
		Token:          token,
		IsContinuation: false,
		ResultCh:       o.resultCh,
	}

	slog.Info("[ORCHESTRATOR] Submitting job to worker pool", "task_id", taskID)
	if err := o.workerPool.Submit(func() {
		o.executeTask(context.Background(), job)
	}); err != nil {
		slog.Error("[ORCHESTRATOR] Failed to submit job to worker pool", "error", err, "task_id", taskID)
		return "", err
	}

	// Publish agent_run_updated event
	o.eventBus.Publish(models.Event{
		TaskID: taskID,
		Type:   string(EventTypeTaskStarted),
		Data:   "",
	})
	slog.Info("[ORCHESTRATOR] Job submitted successfully", "task_id", taskID)

	return taskID, nil
}

// ContinueTask continues a task with a follow-up message.
func (o *Orchestrator) ContinueTask(ctx context.Context, taskID, followUpMsg, agentBackend, provider, model string) error {
	if followUpMsg == "" {
		return fmt.Errorf("follow-up message cannot be empty")
	}

	// Get task info
	task, err := o.repo.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get project info
	var token, owner, repoName string
	if task.RepositoryID != nil {
		repo, err := o.repo.GetRepository(ctx, *task.RepositoryID)
		if err == nil {
			conn, err := o.repo.GetGithubConnectionByID(ctx, repo.ConnectionID)
			if err == nil {
				token = conn.AccessToken
				owner = repo.Owner
				repoName = repo.Name
			}
		}
	}

	// Update status to in_progress
	if err := o.repo.UpdateStatus(ctx, taskID, "in_progress"); err != nil {
		return err
	}

	// Create agent run row
	runID, err := o.repo.CreateAgentRun(ctx, taskID, followUpMsg, agentBackend, provider, model)
	if err != nil {
		return fmt.Errorf("failed to create agent run: %w", err)
	}

	// Append user message to DB immediately
	if err := o.repo.CreateMessage(ctx, taskID, runID, "user", followUpMsg); err != nil {
		slog.Error("[ORCHESTRATOR] Failed to create user message", "error", err)
	}

	// Publish agent_run_updated event
	o.eventBus.Publish(models.Event{
		TaskID: taskID,
		Type:   string(EventTypeAgentRunUpdated),
		Data:   "",
	})

	// Publish task_updated event
	o.eventBus.Publish(models.Event{
		TaskID: taskID,
		Type:   string(EventTypeTaskUpdated),
		Data:   "",
	})

	// Load existing messages for state restoration
	messages, err := o.repo.GetMessagesByTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to load messages: %w", err)
	}

	// Convert to JSON for agent state restoration
	messageHistoryJSON, err := ConvertMessagesToJSON(messages)
	if err != nil {
		return fmt.Errorf("failed to convert messages: %w", err)
	}

	// Submit job to worker pool
	job := TaskJob{
		TaskID:         taskID,
		ProjectID:      *task.RepositoryID,
		Intent:         followUpMsg,
		ModelID:        model,
		Owner:          owner,
		Repo:           repoName,
		Token:          token,
		MessageHistory: messageHistoryJSON,
		IsContinuation: true,
		ResultCh:       o.resultCh,
	}

	if err := o.workerPool.Submit(func() {
		o.executeTask(context.Background(), job)
	}); err != nil {
		return err
	}

	return nil
}

// executeTask executes a single task.
func (o *Orchestrator) executeTask(ctx context.Context, job TaskJob) {
	slog.Info("[ORCHESTRATOR] Executing task", "task_id", job.TaskID, "intent", job.Intent, "continuation", job.IsContinuation)

	// Check if incoming context is already cancelled
	if ctx.Err() != nil {
		slog.Error("[ORCHESTRATOR] Incoming context already cancelled", "task_id", job.TaskID, "error", ctx.Err())
		job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: fmt.Sprintf("context cancelled before execution: %v", ctx.Err())}
		return
	}

	// Create a fresh context with timeout (don't inherit from request context which may be cancelled)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Track running task
	o.mu.Lock()
	o.running[job.TaskID] = cancel
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		delete(o.running, job.TaskID)
		o.mu.Unlock()
	}()

	// Update task to in_progress if not continuation
	if !job.IsContinuation {
		slog.Info("[ORCHESTRATOR] Updating task status to in_progress", "task_id", job.TaskID)
		if err := o.repo.UpdateStatus(ctx, job.TaskID, "in_progress"); err != nil {
			slog.Error("[ORCHESTRATOR] Failed to update status", "error", err)
			job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: err.Error()}
			return
		}
		slog.Info("[ORCHESTRATOR] Task status updated successfully", "task_id", job.TaskID)
	}

	// Publish agent_run_started event
	o.eventBus.Publish(models.Event{
		TaskID: job.TaskID,
		Type:   string(EventTypeAgentRunStarted),
		Data:   "",
	})

	// Create agent run if not continuation
	if !job.IsContinuation {
		slog.Info("[ORCHESTRATOR] Creating agent run", "task_id", job.TaskID, "intent", job.Intent)
		runID, err := o.repo.CreateAgentRun(ctx, job.TaskID, job.Intent, "native", "", "")
		if err != nil {
			slog.Error("[ORCHESTRATOR] Failed to create agent run", "error", err, "task_id", job.TaskID)
			job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: err.Error()}
			return
		}
		slog.Info("[ORCHESTRATOR] Agent run created successfully", "task_id", job.TaskID)

		// Append user message to DB immediately
		if err := o.repo.CreateMessage(ctx, job.TaskID, runID, "user", job.Intent); err != nil {
			slog.Error("[ORCHESTRATOR] Failed to create user message", "error", err)
		}
	}

	// Create worktree for isolated execution
	slog.Info("[ORCHESTRATOR] Creating worktree", "task_id", job.TaskID, "owner", job.Owner, "repo", job.Repo)
	worktreePath, err := o.gitReposManager.CreateWorktree(job.Owner, job.Repo, job.TaskID, "agent/task-"+job.TaskID)
	if err != nil {
		slog.Error("[ORCHESTRATOR] Failed to create worktree", "error", err)
		job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: err.Error()}
		return
	}
	slog.Info("[ORCHESTRATOR] Worktree created", "task_id", job.TaskID, "path", worktreePath)

	// Parse ModelID first to determine provider (format: "provider#model" e.g., "zai#glm-4.7" or "o#anthropic/claude-sonnet-4.5")
	provider := ""
	model := ""
	if job.ModelID != "" {
		parts := strings.SplitN(job.ModelID, "#", 2)
		if len(parts) == 2 {
			providerPrefix := parts[0]
			model = parts[1]
			// Map provider prefix to actual provider name
			switch providerPrefix {
			case "o":
				provider = "openrouter"
			case "zai":
				provider = "zai"
			default:
				provider = providerPrefix
			}
		} else {
			model = parts[0]
		}
	}

	// Get API key for the provider (or default if provider is empty)
	slog.Info("[ORCHESTRATOR] Getting API key from settings", "task_id", job.TaskID, "provider", provider)
	apiKey, actualProvider, actualModel, err := o.settings.GetAPIKeyForProvider(ctx, provider)
	if err != nil {
		slog.Error("[ORCHESTRATOR] Failed to get API key", "error", err)
		job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: err.Error()}
		return
	}
	provider = actualProvider
	if model == "" {
		model = actualModel
	}
	slog.Info("[ORCHESTRATOR] Retrieved API settings", "task_id", job.TaskID, "provider", provider, "model", model)

	// Create LLM provider
	var llmProvider llm.Provider
	switch provider {
	case "anthropic":
		llmProvider = llm.NewAnthropicProvider(apiKey)
	case "openrouter":
		llmProvider = llm.NewOpenRouterProvider(apiKey)
	case "zai":
		llmProvider = llm.NewZaiProvider(apiKey)
	default:
		job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: fmt.Sprintf("unsupported provider: %s", provider)}
		return
	}
	llmProvider.SetModel(model)

	// Create agent backend
	var backend agent.Backend
	backend, err = agent.NewNativeBackend(
		agent.WithProvider(llmProvider),
		agent.WithWorkDir(worktreePath),
		agent.WithCallback(func(e agent.StreamEvent) {
			o.handleAgentEvent(job.TaskID, e)
		}),
	)
	if err != nil {
		slog.Error("[ORCHESTRATOR] Failed to create backend", "error", err)
		job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: err.Error()}
		return
	}

	// Restore state if continuing
	if job.IsContinuation && job.MessageHistory != "" {
		if err := backend.RestoreState(job.MessageHistory); err != nil {
			slog.Error("[ORCHESTRATOR] Failed to restore state", "error", err)
		}
	}

	// Execute task
	slog.Info("[ORCHESTRATOR] Starting agent execution", "task_id", job.TaskID, "is_continuation", job.IsContinuation)
	execErr := backend.Run(ctx, job.Intent)
	if execErr != nil {
		slog.Error("[ORCHESTRATOR] Agent execution failed", "error", execErr, "task_id", job.TaskID)
		job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: execErr.Error()}
		return
	}

	slog.Info("[ORCHESTRATOR] Agent execution completed", "task_id", job.TaskID)

	// Commit changes
	commitMessage := fmt.Sprintf("Task: %s", job.Intent)
	if err := o.gitReposManager.CommitAndPush(job.TaskID, commitMessage); err != nil {
		slog.Error("[ORCHESTRATOR] Failed to commit and push", "error", err)
		// Don't fail task - commit might fail if no changes
	}

	// Get git diff
	gitDiff, err := o.gitReposManager.GetDiff(job.TaskID)
	if err != nil {
		slog.Warn("[ORCHESTRATOR] Failed to get git diff", "task_id", job.TaskID, "error", err)
	}
	if gitDiff != "" {
		slog.Info("[ORCHESTRATOR] Git diff generated", "task_id", job.TaskID, "diff_size", len(gitDiff))
	}

	// Get final message from backend
	finalMessage := backend.FinalMessage()

	// Send result
	job.ResultCh <- TaskResult{
		TaskID:      job.TaskID,
		Success:     true,
		AgentOutput: finalMessage,
		GitDiff:     gitDiff,
	}

	slog.Info("[ORCHESTRATOR] Task completed", "task_id", job.TaskID, "success", true)
}

// handleAgentEvent processes agent events and publishes to UI.
func (o *Orchestrator) handleAgentEvent(taskID string, event agent.StreamEvent) {
	switch event.Type {
	case agent.EventTool:
		o.saveMessage(taskID, agent.Message{Role: "tool", Content: []agent.ContentBlock{{Type: "text", Text: event.Content}}})
	case agent.EventToolResult:
		o.saveMessage(taskID, agent.Message{Role: "tool_result", Content: []agent.ContentBlock{{Type: "text", Text: event.Content}}})
	case agent.EventText:
		o.saveMessage(taskID, agent.Message{Role: "assistant", Content: []agent.ContentBlock{{Type: "text", Text: event.Content}}})
	case agent.EventError:
	case agent.EventDone:
	case agent.EventTodo:
	}

	o.eventBus.Publish(models.Event{TaskID: taskID, Type: string(EventTypeAgentRunUpdated), Data: ""})
}

// saveMessage saves a single agent message to database.
func (o *Orchestrator) saveMessage(taskID string, msg agent.Message) {
	// Get latest run ID
	run, err := o.repo.GetLatestAgentRun(context.Background(), taskID)
	if err != nil {
		slog.Error("[ORCHESTRATOR] Failed to get latest run", "error", err)
		return
	}

	// Extract text content from content blocks
	var textContent string
	for _, block := range msg.Content {
		if block.Type == "text" {
			textContent += block.Text
		}
	}

	slog.Info("[ORCHESTRATOR] Saving message", "task_id", taskID, "run_id", run.ID, "role", msg.Role, "content", textContent)
	if err := o.repo.CreateMessage(context.Background(), taskID, run.ID, msg.Role, textContent); err != nil {
		slog.Error("[ORCHESTRATOR] Failed to save message", "error", err)
	}

	// Publish agent_run_updated
	o.eventBus.Publish(models.Event{TaskID: taskID, Type: string(EventTypeAgentRunUpdated), Data: ""})
}

// processResults processes task results from the worker pool.
func (o *Orchestrator) processResults() {
	for result := range o.resultCh {
		// Update task status based on result
		ctx := context.Background()

		if result.Success {
			if err := o.repo.UpdateStatus(ctx, result.TaskID, "review"); err != nil {
				slog.Error("[ORCHESTRATOR] Failed to update task status", "error", err)
			}
			// Update agent run as completed
			if run, err := o.repo.GetLatestAgentRun(ctx, result.TaskID); err == nil && run != nil {
				if err := o.repo.UpdateAgentRunCompleted(ctx, run.ID); err != nil {
					slog.Error("[ORCHESTRATOR] Failed to update agent run completed", "error", err)
				}
			}
		} else {
			if err := o.repo.UpdateStatus(ctx, result.TaskID, "failed"); err != nil {
				slog.Error("[ORCHESTRATOR] Failed to update task status", "error", err)
			}
		}

		slog.Info("[ORCHESTRATOR] Result processed", "task_id", result.TaskID, "success", result.Success)
	}
}

// MergeTask merges task branch to main and pushes.
func (o *Orchestrator) MergeTask(ctx context.Context, taskID string) error {
	// Get task info
	taskInfo, err := o.repo.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get repo info
	var owner, repoName string
	if taskInfo.RepositoryID != nil {
		repo, err := o.repo.GetRepository(ctx, *taskInfo.RepositoryID)
		if err == nil {
			owner = repo.Owner
			repoName = repo.Name
		}
	}

	// Merge to main
	_, err = o.gitReposManager.MergeToMain(owner, repoName, taskID)
	if err != nil {
		// Check for merge conflict
		if _, isConflict := err.(*ErrMergeConflict); isConflict {
			return err
		}
		return fmt.Errorf("failed to merge: %w", err)
	}

	// Update task status to done
	if err := o.repo.UpdateStatus(ctx, taskID, "done"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Publish task_updated event
	o.eventBus.Publish(models.Event{TaskID: taskID, Type: string(EventTypeTaskUpdated), Data: ""})

	return nil
}

// GetConflictDetails returns conflict details for a task.
func (o *Orchestrator) GetConflictDetails(ctx context.Context, taskID string) ([]ConflictFile, error) {
	// Get worktree path
	worktreePath := o.gitReposManager.WorktreePath(taskID)

	// Read all files in worktree
	var conflicts []ConflictFile

	err := filepath.Walk(worktreePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			return nil // skip directories
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // skip files that can't be read
		}

		// Check for conflict markers
		if strings.Contains(string(content), "<<<<<<<") {
			relPath, _ := filepath.Rel(worktreePath, path)
			conflictFile, err := parseConflictFile(path, string(content))
			if err == nil {
				conflictFile.Path = relPath
				conflicts = append(conflicts, *conflictFile)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan for conflicts: %w", err)
	}

	return conflicts, nil
}

// ResolveConflict resolves a merge conflict for a specific file.
func (o *Orchestrator) ResolveConflict(ctx context.Context, taskID, filePath, resolution string) error {
	// Get worktree path
	worktreePath := o.gitReposManager.WorktreePath(taskID)
	fullPath := filepath.Join(worktreePath, filePath)

	// Write resolved content
	if err := os.WriteFile(fullPath, []byte(resolution), 0644); err != nil {
		return fmt.Errorf("failed to write resolution: %w", err)
	}

	return nil
}

// CompleteMergeResolution completes merge conflict resolution and merges to main.
func (o *Orchestrator) CompleteMergeResolution(ctx context.Context, taskID string) error {
	// Get task info
	task, err := o.repo.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Get repo info
	var owner, repoName string
	if task.RepositoryID != nil {
		repo, err := o.repo.GetRepository(ctx, *task.RepositoryID)
		if err == nil {
			owner = repo.Owner
			repoName = repo.Name
		}
	}

	// Commit resolution
	if err := o.gitReposManager.CommitMergeResolution(taskID, "Resolved merge conflicts"); err != nil {
		return fmt.Errorf("failed to commit resolution: %w", err)
	}

	// Merge to main
	if _, err := o.gitReposManager.MergeToMain(owner, repoName, taskID); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	// Update task status to done
	if err := o.repo.UpdateStatus(ctx, taskID, "done"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Publish task_updated event
	o.eventBus.Publish(models.Event{TaskID: taskID, Type: string(EventTypeTaskUpdated), Data: ""})

	return nil
}

// AbortMerge aborts an in-progress merge.
func (o *Orchestrator) AbortMerge(ctx context.Context, taskID string) error {
	return o.gitReposManager.AbortMerge(taskID)
}

// CreatePR creates a pull request for a task.
func (o *Orchestrator) CreatePR(ctx context.Context, taskID string) (string, error) {
	// Get task info
	task, err := o.repo.Get(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("task not found: %w", err)
	}

	// Get project info
	repos, err := o.github.GetRepos(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	var owner, repoName string
	for _, p := range repos {
		if p.ID == *task.RepositoryID {
			owner = p.Owner
			repoName = p.Name
			break
		}
	}

	if owner == "" {
		return "", fmt.Errorf("project not found for task")
	}

	// Get branch name from worktree
	branchName, err := o.gitReposManager.GetCurrentBranch(taskID)
	if err != nil {
		return "", fmt.Errorf("failed to get branch name: %w", err)
	}
	branchName = strings.TrimSpace(branchName)

	// Push branch to remote before creating PR
	if err := o.gitReposManager.PushBranch(taskID); err != nil {
		return "", fmt.Errorf("failed to push branch: %w", err)
	}

	// Create PR
	prURL, err := o.github.CreatePullRequest(ctx, owner, repoName, branchName, task.Title, task.Intent)
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %w", err)
	}

	slog.Info("[ORCHESTRATOR] Created PR", "task_id", taskID, "pr_url", prURL)

	// Update task status to done
	if err := o.repo.UpdateStatus(ctx, taskID, "done"); err != nil {
		return prURL, fmt.Errorf("PR created but failed to update task status: %w", err)
	}

	// Emit status change event
	o.eventBus.Publish(models.Event{TaskID: taskID, Type: string(EventTypeTaskUpdated), Data: ""})

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
	return o.gitReposManager.RemoveWorktree(taskID)
}

// SearchProjectFiles searches for files in a project using fuzzy matching.
// Returns a list of file paths relative to the repo root, sorted by match score.
func (o *Orchestrator) SearchProjectFiles(ctx context.Context, projectID, query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 20
	}

	// Get project info
	var owner, repoName string
	if projectID != "" {
		repo, err := o.repo.GetRepository(ctx, projectID)
		if err == nil {
			owner = repo.Owner
			repoName = repo.Name
		}
	}

	if owner == "" {
		return nil, fmt.Errorf("project not found")
	}

	repoPath := o.gitReposManager.RepoPath(owner, repoName)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository not cloned yet")
	}

	// Collect all file paths
	var files []string
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
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

// truncate truncates a string to a maximum length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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
// Higher score = better match. 0 = no match.
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

// parseConflictFile parses a file with git conflict markers.
func parseConflictFile(path, content string) (*ConflictFile, error) {
	lines := strings.Split(content, "\n")

	var (
		base     strings.Builder
		current  strings.Builder
		incoming strings.Builder
		section  int // 0=before, 1=current, 2=incoming
	)

	for _, line := range lines {
		if strings.HasPrefix(line, "<<<<<<<") {
			section = 1
			current.WriteString(line + "\n")
		} else if strings.HasPrefix(line, "=======") {
			section = 2
			incoming.WriteString(line + "\n")
		} else if strings.HasPrefix(line, ">>>>>>>") {
			section = 0
		} else {
			switch section {
			case 0:
				base.WriteString(line + "\n")
			case 1:
				current.WriteString(line + "\n")
			case 2:
				incoming.WriteString(line + "\n")
			}
		}
	}

	return &ConflictFile{
		Path:     path,
		Base:     base.String(),
		Current:  current.String(),
		Incoming: incoming.String(),
	}, nil
}
