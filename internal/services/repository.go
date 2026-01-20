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

// Repository handles task persistence.
type Repository struct {
	db *db.DB
}

// NewRepository creates a new task service.
func NewRepository(database *db.DB) *Repository {
	return &Repository{db: database}
}

func (s *Repository) GetRepository(ctx context.Context, projectID string) (sqlc.Repository, error) {
	return s.db.Queries.GetRepository(ctx, projectID)
}

func (s *Repository) GetGithubConnectionByID(ctx context.Context, githubConnectionID string) (sqlc.GithubConnection, error) {
	return s.db.Queries.GetGithubConnectionByID(ctx, githubConnectionID)

}

// Create creates a new task with validation.
func (s *Repository) Create(ctx context.Context, repositoryID, intent string) (*models.Task, error) {
	id := shortuuid.New()
	// Validate input
	if intent == "" {
		return nil, fmt.Errorf("intent is required")
	}

	now := time.Now().UnixMilli()
	if err := s.db.Queries.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:           id,
		RepositoryID: sql.NullString{String: repositoryID, Valid: repositoryID != ""},
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
func (s *Repository) Get(ctx context.Context, id string) (*models.Task, error) {
	task, err := s.db.Queries.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return sqlcTaskToModel(&task), nil
}

// List retrieves all tasks.
func (s *Repository) List(ctx context.Context) ([]*models.Task, error) {
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
func (s *Repository) ListByStatus(ctx context.Context, status string) ([]*models.Task, error) {
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
func (s *Repository) UpdateStatus(ctx context.Context, id, status string) error {
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
func (s *Repository) Delete(ctx context.Context, id string) error {
	if err := s.db.Queries.DeleteTask(ctx, id); err != nil {
		return err
	}
	return nil
}

// GetPendingTasks retrieves all pending tasks for execution.
func (s *Repository) GetPendingTasks(ctx context.Context) ([]*models.Task, error) {
	return s.ListByStatus(ctx, "pending")
}

// GetInProgressTasks retrieves all in-progress tasks.
func (s *Repository) GetInProgressTasks(ctx context.Context) ([]*models.Task, error) {
	return s.ListByStatus(ctx, "in_progress")
}

// sqlcTaskToModel converts sqlc task to model.
func sqlcTaskToModel(task *sqlc.Task) *models.Task {
	return &models.Task{
		ID:           task.ID,
		RepositoryID: nullableString(task.RepositoryID),
		Title:        task.Title,
		Intent:       task.Intent,
		Status:       task.Status,
		Position:     nullableInt64(task.Position),
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}

// nullableInt64 converts sql.NullInt64 to *int64.
func nullableInt64(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

// nullableString converts sql.NullString to *string.
func nullableString(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}
