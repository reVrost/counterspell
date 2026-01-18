package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lithammer/shortuuid/v4"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
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
func (s *TaskService) Create(ctx context.Context, userID, projectID, title, intent string) (*models.Task, error) {
	id := shortuuid.New()
	now := time.Now()

	task, err := s.db.Queries.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:        id,
		UserID:    userID,
		ProjectID: projectID,
		Title:     title,
		Intent:    intent,
		Status:    string(models.StatusPending),
		Position:  pgtype.Int4{Int32: 0, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return sqlcTaskToModel(task), nil
}

// Get retrieves a task by ID.
func (s *TaskService) Get(ctx context.Context, userID, id string) (*models.Task, error) {
	task, err := s.db.Queries.GetTask(ctx, sqlc.GetTaskParams{
		ID:     id,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return sqlcTaskToModel(task), nil
}

// GetWithProject retrieves a task by ID with project info in one query.
func (s *TaskService) GetWithProject(ctx context.Context, userID, id string) (*models.TaskWithProject, error) {
	row, err := s.db.Queries.GetTaskWithProject(ctx, sqlc.GetTaskWithProjectParams{
		ID:     id,
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task with project: %w", err)
	}

	task := &models.Task{
		ID:        row.ID,
		ProjectID: row.ProjectID,
		Title:     row.Title,
		Intent:    row.Intent,
		Status:    models.TaskStatus(row.Status),
		Position:  int(row.Position.Int32),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
	if row.CurrentStep.Valid {
		task.CurrentStep = row.CurrentStep.String
	}
	if row.AssignedAgentID.Valid {
		task.AssignedAgentID = row.AssignedAgentID.String
	}
	if row.AssignedUserID.Valid {
		task.AssignedUserID = row.AssignedUserID.String
	}

	var project *models.Project
	if row.ProjectOwner.Valid && row.ProjectRepo.Valid {
		project = &models.Project{
			ID:          row.ProjectID,
			GitHubOwner: row.ProjectOwner.String,
			GitHubRepo:  row.ProjectRepo.String,
		}
	}

	return &models.TaskWithProject{
		Task:    task,
		Project: project,
	}, nil
}

// List returns all tasks, optionally filtered by status and/or project.
func (s *TaskService) List(ctx context.Context, userID string, status *models.TaskStatus, projectID *string) ([]models.Task, error) {
	var tasks []sqlc.Task
	var err error

	switch {
	case status != nil && projectID != nil:
		tasks, err = s.db.Queries.ListTasksByStatusAndProject(ctx, sqlc.ListTasksByStatusAndProjectParams{
			UserID:    userID,
			Status:    string(*status),
			ProjectID: *projectID,
		})
	case status != nil:
		tasks, err = s.db.Queries.ListTasksByStatus(ctx, sqlc.ListTasksByStatusParams{
			UserID: userID,
			Status: string(*status),
		})
	case projectID != nil:
		tasks, err = s.db.Queries.ListTasksByProject(ctx, sqlc.ListTasksByProjectParams{
			UserID:    userID,
			ProjectID: *projectID,
		})
	default:
		tasks, err = s.db.Queries.ListTasks(ctx, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	result := make([]models.Task, len(tasks))
	for i, t := range tasks {
		result[i] = *sqlcTaskToModel(t)
	}
	return result, nil
}

// UpdateStatus updates a task's status.
func (s *TaskService) UpdateStatus(ctx context.Context, userID, id string, status models.TaskStatus) error {
	err := s.db.Queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
		Status:    string(status),
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ID:        id,
		UserID:    userID,
	})
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

// UpdateStep updates a task's current step.
func (s *TaskService) UpdateStep(ctx context.Context, userID, id, step string) error {
	err := s.db.Queries.UpdateTaskStep(ctx, sqlc.UpdateTaskStepParams{
		CurrentStep: pgtype.Text{String: step, Valid: step != ""},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ID:          id,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("failed to update task step: %w", err)
	}
	return nil
}

// UpdateStatusAndStep updates both status and current step.
func (s *TaskService) UpdateStatusAndStep(ctx context.Context, userID, id string, status models.TaskStatus, step string) error {
	err := s.db.Queries.UpdateTaskStatusAndStep(ctx, sqlc.UpdateTaskStatusAndStepParams{
		Status:      string(status),
		CurrentStep: pgtype.Text{String: step, Valid: step != ""},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ID:          id,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("failed to update task status and step: %w", err)
	}
	return nil
}

// Move updates a task's position within its status column.
func (s *TaskService) Move(ctx context.Context, userID, id string, newPosition int) error {
	err := s.db.Queries.UpdateTaskPosition(ctx, sqlc.UpdateTaskPositionParams{
		Position:  pgtype.Int4{Int32: int32(newPosition), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ID:        id,
		UserID:    userID,
	})
	if err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}
	return nil
}

// UpdatePositionAndStatus updates both position and status in one operation.
func (s *TaskService) UpdatePositionAndStatus(ctx context.Context, userID, id string, status models.TaskStatus, position int) error {
	err := s.db.Queries.UpdateTaskPositionAndStatus(ctx, sqlc.UpdateTaskPositionAndStatusParams{
		Status:    string(status),
		Position:  pgtype.Int4{Int32: int32(position), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ID:        id,
		UserID:    userID,
	})
	if err != nil {
		return fmt.Errorf("failed to update task position and status: %w", err)
	}
	return nil
}

// Delete removes a task.
func (s *TaskService) Delete(ctx context.Context, userID, id string) error {
	err := s.db.Queries.DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// sqlcTaskToModel converts a sqlc Task to a models.Task
func sqlcTaskToModel(t sqlc.Task) *models.Task {
	task := &models.Task{
		ID:        t.ID,
		ProjectID: t.ProjectID,
		Title:     t.Title,
		Intent:    t.Intent,
		Status:    models.TaskStatus(t.Status),
		Position:  int(t.Position.Int32),
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
	}
	if t.CurrentStep.Valid {
		task.CurrentStep = t.CurrentStep.String
	}
	if t.AssignedAgentID.Valid {
		task.AssignedAgentID = t.AssignedAgentID.String
	}
	if t.AssignedUserID.Valid {
		task.AssignedUserID = t.AssignedUserID.String
	}
	return task
}
