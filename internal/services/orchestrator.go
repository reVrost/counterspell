package services

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lithammer/shortuuid/v4"
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
	gitReposManager  *GitManager
	eventBus         *EventBus
	settings         *SettingsService
	github           *GitHubService
	dataDir          string
	workerPool       *ants.Pool
	resultCh         chan TaskResult
	running          map[string]context.CancelFunc
	mu               sync.Mutex
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
		repo:             repo,
		eventBus:          eventBus,
		settings:          settings,
		github:            github,

		gitReposManager:  NewGitManager(dataDir),
		dataDir:           dataDir,

		// Worker related fields
		workerPool:        pool,
		resultCh:          make(chan TaskResult, 100),
		running:           make(map[string]context.CancelFunc),
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
	o.workerPool.Release()
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

	taskID := shortuuid.New()

	// Create task in database
	_, err = o.repo.Create(ctx, projectID, intent)
	if err != nil {
		return "", err
	}

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
		ResultCh:        o.resultCh,
	}

	if err := o.workerPool.Submit(func() {
		o.executeTask(ctx, job)
	}); err != nil {
		return "", err
	}

	return taskID, nil
}

// ContinueTask continues a task with a follow-up message.
func (o *Orchestrator) ContinueTask(ctx context.Context, taskID, followUpMessage, modelID string) error {
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
	_, err = o.repo.CreateAgentRun(ctx, taskID, followUpMessage, "native", "", "")
	if err != nil {
		return fmt.Errorf("failed to create agent run: %w", err)
	}

	// Append user message to DB immediately
	if err := o.repo.CreateMessage(ctx, taskID, "", "user", followUpMessage, "", ""); err != nil {
		slog.Error("[ORCHESTRATOR] Failed to create user message", "error", err)
	}

	// Publish agent_run_updated event
	o.eventBus.Publish(models.Event{
		TaskID: taskID,
		Type:    "agent_run_updated",
		Data:    "",
	})

	// Publish task_updated event
	o.eventBus.Publish(models.Event{
		TaskID: taskID,
		Type:    "task_updated",
		Data:    "",
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
		Intent:         followUpMessage,
		ModelID:        modelID,
		Owner:          owner,
		Repo:           repoName,
		Token:          token,
		MessageHistory: messageHistoryJSON,
		IsContinuation: true,
		ResultCh:        o.resultCh,
	}

	if err := o.workerPool.Submit(func() {
		o.executeTask(ctx, job)
	}); err != nil {
		return err
	}

	return nil
}

// executeTask executes a single task.
func (o *Orchestrator) executeTask(ctx context.Context, job TaskJob) {
	slog.Info("[ORCHESTRATOR] Executing task", "task_id", job.TaskID, "intent", job.Intent, "continuation", job.IsContinuation)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
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
		if err := o.repo.UpdateStatus(ctx, job.TaskID, "in_progress"); err != nil {
			slog.Error("[ORCHESTRATOR] Failed to update status", "error", err)
		}
	}

	// Publish agent_run_started event
	o.eventBus.Publish(models.Event{
		TaskID: job.TaskID,
		Type:    "agent_run_started",
		Data:    "",
	})

	// Create agent run if not continuation
	if !job.IsContinuation {
		_, err := o.repo.CreateAgentRun(ctx, job.TaskID, job.Intent, "native", "", "")
		if err != nil {
			slog.Error("[ORCHESTRATOR] Failed to create agent run", "error", err)
			job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: err.Error()}
			return
		}
	}

	// Create worktree for isolated execution
	worktreePath, err := o.gitReposManager.CreateWorktree(job.Owner, job.Repo, job.TaskID, "agent/task-"+job.TaskID)
	if err != nil {
		slog.Error("[ORCHESTRATOR] Failed to create worktree", "error", err)
		job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: err.Error()}
		return
	}
	o.emit(job.TaskID, "plan", fmt.Sprintf("Created worktree at %s", worktreePath))

	// Get settings for API key and provider
	apiKey, provider, model, err := o.settings.GetAPIKey(ctx)
	if err != nil {
		slog.Error("[ORCHESTRATOR] Failed to get API key", "error", err)
		job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: err.Error()}
		return
	}

	// Parse ModelID if provided
	if job.ModelID != "" {
		parts := strings.SplitN(job.ModelID, ":", 2)
		if len(parts) == 2 {
			provider = parts[0]
			model = parts[1]
		} else {
			model = parts[0]
		}
	}

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
	var execErr error
	if job.IsContinuation {
		execErr = backend.Send(ctx, job.Intent)
	} else {
		execErr = backend.Run(ctx, job.Intent)
	}

	if execErr != nil {
		slog.Error("[ORCHESTRATOR] Agent execution failed", "error", execErr)
		o.emitError(job.TaskID, fmt.Sprintf("Agent execution failed: %s", execErr.Error()))
		job.ResultCh <- TaskResult{TaskID: job.TaskID, Success: false, Error: execErr.Error()}
		return
	}

	// Commit changes
	commitMessage := fmt.Sprintf("Task: %s", job.Intent)
	if err := o.gitReposManager.CommitAndPush(job.TaskID, commitMessage); err != nil {
		slog.Error("[ORCHESTRATOR] Failed to commit and push", "error", err)
		o.emitError(job.TaskID, fmt.Sprintf("Failed to commit: %s", err.Error()))
		// Don't fail task - commit might fail if no changes
	}

	// Get git diff
	gitDiff, _ := o.gitReposManager.GetDiff(job.TaskID)
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
		// Save message history to DB and publish agent_run_updated
		if err := o.saveMessageHistory(taskID, event.Messages); err != nil {
			slog.Error("[ORCHESTRATOR] Failed to save message history", "error", err)
		}
		o.eventBus.Publish(models.Event{TaskID: taskID, Type: "agent_run_updated", Data: ""})
	case agent.EventTodo:
		o.eventBus.Publish(models.Event{TaskID: taskID, Type: "agent_run_updated", Data: ""})
	}
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
				o.repo.UpdateAgentRunCompleted(ctx, run.ID)
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
	o.eventBus.Publish(models.Event{TaskID: taskID, Type: "task_updated", Data: ""})

	return nil
}

// GetConflictDetails returns conflict details for a task.
func (o *Orchestrator) GetConflictDetails(ctx context.Context, taskID string) ([]ConflictFile, error) {
	// Get worktree path
	worktreePath := o.gitReposManager.WorktreePath(taskID)

	// Read all files in worktree
	var conflicts []ConflictFile
	var err error
	err = filepath.Walk(worktreePath, func(path string, info os.FileInfo, err error) error {
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

	// Stage file
	o.emit(taskID, "info", fmt.Sprintf("Resolved conflict in %s and staged", filePath))

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
	o.eventBus.Publish(models.Event{TaskID: taskID, Type: "task_updated", Data: ""})

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
	o.eventBus.Publish(models.Event{TaskID: taskID, Type: "task_updated", Data: ""})

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

// saveMessageHistory saves agent message history to database.
func (o *Orchestrator) saveMessageHistory(taskID, messagesJSON string) error {
	// Parse messages JSON
	var messages []struct {
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text,omitempty"`
		} `json:"content"`
	}

	if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
		return fmt.Errorf("failed to parse messages: %w", err)
	}

	// Get latest run ID
	run, err := o.repo.GetLatestAgentRun(context.Background(), taskID)
	if err != nil {
		return fmt.Errorf("failed to get latest run: %w", err)
	}

	// Save each message
	ctx := context.Background()
	for _, msg := range messages {
		// Skip system messages for now
		if msg.Role == "system" {
			continue
		}

		// Extract text content from content blocks
		var textContent string
		for _, block := range msg.Content {
			if block.Type == "text" {
				textContent += block.Text
			}
		}

		if err := o.repo.CreateMessage(ctx, taskID, run.ID, msg.Role, textContent, "", ""); err != nil {
			slog.Error("[ORCHESTRATOR] Failed to save message", "error", err)
		}
	}

	return nil
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

	o.eventBus.Publish(models.Event{
		TaskID:      taskID,
		Type:        "log",
		Data:        htmlPayload,
	})
}

// emitError sends an error event to the UI.
func (o *Orchestrator) emitError(taskID, message string) {
	o.emit(taskID, "error", message)
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
