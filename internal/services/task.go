package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/models"
)

// TaskService handles task persistence and business logic.
type TaskService struct {
	db *db.DB
}

// NewTaskService creates a new task service.
func NewTaskService(database *db.DB) *TaskService {
	return &TaskService{db: database}
}

// Create creates a new task.
func (s *TaskService) Create(ctx context.Context, title, intent string) (*models.Task, error) {
	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO tasks (id, title, intent, status, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query, id, title, intent, models.StatusTodo, 0, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return s.Get(ctx, id)
}

// Get retrieves a task by ID.
func (s *TaskService) Get(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, title, intent, status, position, created_at, updated_at
		FROM tasks WHERE id = ?
	`
	var task models.Task
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Intent, &task.Status,
		&task.Position, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// List returns all tasks, optionally filtered by status.
func (s *TaskService) List(ctx context.Context, status *models.TaskStatus) ([]models.Task, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `
			SELECT id, title, intent, status, position, created_at, updated_at
			FROM tasks WHERE status = ?
			ORDER BY position ASC, created_at DESC
		`
		args = []interface{}{*status}
	} else {
		query = `
			SELECT id, title, intent, status, position, created_at, updated_at
			FROM tasks
			ORDER BY status ASC, position ASC, created_at DESC
		`
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Intent, &task.Status,
			&task.Position, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateStatus updates a task's status.
func (s *TaskService) UpdateStatus(ctx context.Context, id string, status models.TaskStatus) error {
	query := `
		UPDATE tasks SET status = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

// Move updates a task's position within its status column.
func (s *TaskService) Move(ctx context.Context, id string, newPosition int) error {
	query := `
		UPDATE tasks SET position = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, newPosition, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}
	return nil
}

// UpdatePositionAndStatus updates both position and status in one operation.
func (s *TaskService) UpdatePositionAndStatus(ctx context.Context, id string, status models.TaskStatus, position int) error {
	query := `
		UPDATE tasks SET status = ?, position = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, status, position, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update task position and status: %w", err)
	}
	return nil
}

// Delete removes a task.
func (s *TaskService) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// ResetInProgress resets all tasks stuck in in_progress back to todo.
// This is called on server startup for recovery.
func (s *TaskService) ResetInProgress(ctx context.Context) error {
	query := `
		UPDATE tasks SET status = ?, updated_at = ?
		WHERE status = ?
	`
	result, err := s.db.ExecContext(ctx, query, models.StatusTodo, time.Now(), models.StatusInProgress)
	if err != nil {
		return fmt.Errorf("failed to reset in-progress tasks: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows > 0 {
		slog.Info("Reset stuck tasks", "count", rows)
	}
	return nil
}
