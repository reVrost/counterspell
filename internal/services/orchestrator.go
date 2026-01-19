package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"github.com/panjf2000/ants/v2"
	"github.com/revrost/code/counterspell/internal/db"
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
	Intent    string
	ModelID   string
	ProjectID string
	RunID     string // Current agent run ID
	ResultCh  chan<- TaskResult
}

// Orchestrator manages task execution with agents.
type Orchestrator struct {
	db         *db.DB
	tasks      *TaskService
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
	repos *RepoManager,
	events *EventBus,
	settings *SettingsService,
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
		db:         database,
		tasks:      tasks,
		repos:      repos,
		events:     events,
		settings:   settings,
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
func (o *Orchestrator) StartTask(ctx context.Context, projectID, intent, modelID string) (string, error) {
	// For local-first mode, projectID can be a local repository path
	taskID := shortuuid.New()
	machineID := o.userID // Default to user ID as machine ID

	// Create task in database
	_, err := o.tasks.Create(ctx, machineID, projectID, intent)
	if err != nil {
		return "", err
	}

	slog.Info("[ORCHESTRATOR] Task created", "task_id", taskID, "project_id", projectID, "intent", intent)

	// Submit job to worker pool
	job := TaskJob{
		TaskID:    taskID,
		UserID:    o.userID,
		Intent:    intent,
		ModelID:   modelID,
		ProjectID: projectID,
		RunID:     taskID,
		ResultCh:  o.resultCh,
	}

	if err := o.workerPool.Submit(func() {
		o.executeTask(ctx, job)
	}); err != nil {
		return "", err
	}

	return taskID, nil
}

// executeTask executes a single task.
func (o *Orchestrator) executeTask(ctx context.Context, job TaskJob) {
	slog.Info("[ORCHESTRATOR] Executing task", "task_id", job.TaskID, "intent", job.Intent)

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

	// Simple task execution (placeholder for agent logic)
	// In a full implementation, this would call to agent backend
	slog.Info("[ORCHESTRATOR] Task execution started", "task_id", job.TaskID)

	// Update task to in_progress
	_ = o.tasks.UpdateStatus(ctx, job.TaskID, "in_progress")

	// TODO: Implement actual agent execution here
	// For now, just mark as done after a short delay
	time.Sleep(1 * time.Second)

	// Send result
	job.ResultCh <- TaskResult{
		TaskID:  job.TaskID,
		UserID:  job.UserID,
		RunID:   job.RunID,
		Success:  true,
		Output:   "Task completed",
	}

	slog.Info("[ORCHESTRATOR] Task completed", "task_id", job.TaskID, "success", true)
}

// processResults processes task results from the worker pool.
func (o *Orchestrator) processResults() {
	for result := range o.resultCh {
		// Update task status based on result
		ctx := context.Background()

		if result.Success {
			if err := o.tasks.UpdateStatus(ctx, result.TaskID, "done"); err != nil {
				slog.Error("[ORCHESTRATOR] Failed to update task status", "error", err)
			}
		} else {
			if err := o.tasks.UpdateStatus(ctx, result.TaskID, "failed"); err != nil {
				slog.Error("[ORCHESTRATOR] Failed to update task status", "error", err)
			}
		}

		slog.Info("[ORCHESTRATOR] Result processed", "task_id", result.TaskID, "success", result.Success)
	}
}

// Stub methods for GitHub-specific features (not used in local-first mode)
func (o *Orchestrator) MergeTask(ctx context.Context, taskID string) error {
	return fmt.Errorf("merge not supported in local-first mode")
}

func (o *Orchestrator) GetConflictDetails(ctx context.Context, taskID string) ([]ConflictFile, error) {
	return nil, fmt.Errorf("conflict resolution not supported in local-first mode")
}

func (o *Orchestrator) ResolveConflict(ctx context.Context, taskID, file, resolution string) error {
	return fmt.Errorf("conflict resolution not supported in local-first mode")
}

func (o *Orchestrator) AbortMerge(ctx context.Context, taskID string) error {
	return fmt.Errorf("merge abort not supported in local-first mode")
}

func (o *Orchestrator) CompleteMergeResolution(ctx context.Context, taskID string) error {
	return fmt.Errorf("merge resolution not supported in local-first mode")
}

func (o *Orchestrator) CreatePR(ctx context.Context, taskID, title, body string) error {
	return fmt.Errorf("PR creation not supported in local-first mode")
}

func (o *Orchestrator) CleanupTask(ctx context.Context, taskID string) error {
	return o.tasks.Delete(ctx, taskID)
}
