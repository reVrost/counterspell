package services

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
	"github.com/revrost/code/counterspell/internal/models"
)

// TaskService handles task persistence.
type TaskService struct {
	db *db.DB
}

// NewTaskService creates a new task service.
func NewTaskService(database *db.DB) *TaskService {
	return &TaskService{db: database}
}

// Create creates a new task with validation.
func (s *TaskService) Create(ctx context.Context, machineID, projectID, intent string) (*models.Task, error) {
	id := shortuuid.New()
	// Validate input
	if machineID == "" {
		return nil, fmt.Errorf("machine_id is required")
	}
	if intent == "" {
		return nil, fmt.Errorf("intent is required")
	}

	now := time.Now().UnixMilli()
	if err := s.db.Queries.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:           id,
		RepositoryID: sql.NullString{String: projectID, Valid: projectID != ""},
		Title:        intent, // Use intent as title for now
		Intent:       intent,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		return nil, err
	}

	return s.Get(ctx, id)
}

// Get retrieves a task by ID.
func (s *TaskService) Get(ctx context.Context, id string) (*models.Task, error) {
	task, err := s.db.Queries.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return sqlcTaskToModel(&task), nil
}

// List retrieves all tasks.
func (s *TaskService) List(ctx context.Context) ([]*models.Task, error) {
	tasks, err := s.db.Queries.ListTasks(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Task, len(tasks))
	for i := range tasks {
		result[i] = sqlcTaskToModel(&tasks[i])
	}
	return result, nil
}

// ListByStatus retrieves tasks by status.
func (s *TaskService) ListByStatus(ctx context.Context, status string) ([]*models.Task, error) {
	tasks, err := s.db.Queries.ListTasksByStatus(ctx, status)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Task, len(tasks))
	for i := range tasks {
		result[i] = sqlcTaskToModel(&tasks[i])
	}
	return result, nil
}

// UpdateStatus updates task status with validation.
func (s *TaskService) UpdateStatus(ctx context.Context, id, status string) error {
	// Validate status
	validStatuses := []string{"pending", "in_progress", "review", "done", "failed"}
	if !slices.Contains(validStatuses, status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	if err := s.db.Queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
		Status: status,
		ID:     id,
	}); err != nil {
		return err
	}
	return nil
}

// Delete removes a task.
func (s *TaskService) Delete(ctx context.Context, id string) error {
	if err := s.db.Queries.DeleteTask(ctx, id); err != nil {
		return err
	}
	return nil
}

// GetPendingTasks retrieves all pending tasks for execution.
func (s *TaskService) GetPendingTasks(ctx context.Context) ([]*models.Task, error) {
	return s.ListByStatus(ctx, "pending")
}

// GetInProgressTasks retrieves all in-progress tasks.
func (s *TaskService) GetInProgressTasks(ctx context.Context) ([]*models.Task, error) {
	return s.ListByStatus(ctx, "in_progress")
}

// sqlcTaskToModel converts sqlc task to model.
func sqlcTaskToModel(task *sqlc.Task) *models.Task {
	var repoID *string
	if task.RepositoryID.Valid {
		repoID = &task.RepositoryID.String
	}
	return &models.Task{
		ID:        task.ID,
		ProjectID: repoID,
		Title:     task.Title,
		Intent:    task.Intent,
		Status:    task.Status,
		Position:  task.Position.Int64,
		CreatedAt: time.UnixMilli(task.CreatedAt),
		UpdatedAt: time.UnixMilli(task.UpdatedAt),
	}
}
