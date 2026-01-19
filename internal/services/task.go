package services

import (
	"context"
	"database/sql"
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

// Create creates a new task.
func (s *TaskService) Create(ctx context.Context, machineID, title, intent string) (*models.Task, error) {
	id := shortuuid.New()
	now := time.Now()

	if err := s.db.Queries.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:        id,
		MachineID: machineID,
		Title:     title,
		Intent:    intent,
		Status:    "pending",
		Position:   sql.NullInt64{Int64: 0, Valid: true},
		CreatedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: sql.NullTime{Time: now, Valid: true},
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

// UpdateStatus updates task status.
func (s *TaskService) UpdateStatus(ctx context.Context, id, status string) error {
	now := time.Now()
	if err := s.db.Queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
		Status:    status,
		UpdatedAt: sql.NullTime{Time: now, Valid: true},
		ID:        id,
	}); err != nil {
		return err
	}
	return nil
}

// UpdateStep updates current step of a task.
func (s *TaskService) UpdateStep(ctx context.Context, id, step string) error {
	now := time.Now()
	if err := s.db.Queries.UpdateTaskStep(ctx, sqlc.UpdateTaskStepParams{
		CurrentStep: sql.NullString{String: step, Valid: step != ""},
		UpdatedAt:   sql.NullTime{Time: now, Valid: true},
		ID:          id,
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

// sqlcTaskToModel converts sqlc task to model.
func sqlcTaskToModel(task *sqlc.Task) *models.Task {
	return &models.Task{
		ID:             task.ID,
		MachineID:      task.MachineID,
		Title:          task.Title,
		Intent:         task.Intent,
		Status:         task.Status,
		CurrentStep:    task.CurrentStep.String,
		AssignedAgentID: task.AssignedAgentID.String,
		CreatedAt:      task.CreatedAt.Time,
		UpdatedAt:      task.UpdatedAt.Time,
	}
}
