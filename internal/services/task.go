package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

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
func (s *TaskService) Create(ctx context.Context, projectID, title, intent string) (*models.Task, error) {
	id := shortuuid.New()
	now := time.Now()

	task, err := s.db.Queries.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:        id,
		ProjectID: projectID,
		Title:     title,
		Intent:    intent,
		Status:    string(models.StatusTodo),
		Position:  sql.NullInt64{Int64: 0, Valid: true},
		CreatedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: sql.NullTime{Time: now, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return sqlcTaskToModel(task), nil
}

// Get retrieves a task by ID.
func (s *TaskService) Get(ctx context.Context, id string) (*models.Task, error) {
	task, err := s.db.Queries.GetTask(ctx, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return sqlcTaskToModel(task), nil
}

// List returns all tasks, optionally filtered by status and/or project.
func (s *TaskService) List(ctx context.Context, status *models.TaskStatus, projectID *string) ([]models.Task, error) {
	var tasks []sqlc.Task
	var err error

	switch {
	case status != nil && projectID != nil:
		tasks, err = s.db.Queries.ListTasksByStatusAndProject(ctx, sqlc.ListTasksByStatusAndProjectParams{
			Status:    string(*status),
			ProjectID: *projectID,
		})
	case status != nil:
		tasks, err = s.db.Queries.ListTasksByStatus(ctx, string(*status))
	case projectID != nil:
		tasks, err = s.db.Queries.ListTasksByProject(ctx, *projectID)
	default:
		tasks, err = s.db.Queries.ListTasks(ctx)
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
func (s *TaskService) UpdateStatus(ctx context.Context, id string, status models.TaskStatus) error {
	err := s.db.Queries.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
		Status:    string(status),
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        id,
	})
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

// Move updates a task's position within its status column.
func (s *TaskService) Move(ctx context.Context, id string, newPosition int) error {
	err := s.db.Queries.UpdateTaskPosition(ctx, sqlc.UpdateTaskPositionParams{
		Position:  sql.NullInt64{Int64: int64(newPosition), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        id,
	})
	if err != nil {
		return fmt.Errorf("failed to move task: %w", err)
	}
	return nil
}

// UpdatePositionAndStatus updates both position and status in one operation.
func (s *TaskService) UpdatePositionAndStatus(ctx context.Context, id string, status models.TaskStatus, position int) error {
	err := s.db.Queries.UpdateTaskPositionAndStatus(ctx, sqlc.UpdateTaskPositionAndStatusParams{
		Status:    string(status),
		Position:  sql.NullInt64{Int64: int64(position), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        id,
	})
	if err != nil {
		return fmt.Errorf("failed to update task position and status: %w", err)
	}
	return nil
}

// Delete removes a task.
func (s *TaskService) Delete(ctx context.Context, id string) error {
	err := s.db.Queries.DeleteTask(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// UpdateWithResult updates a task's status, agent output, git diff, and message history.
func (s *TaskService) UpdateWithResult(ctx context.Context, id string, status models.TaskStatus, agentOutput, gitDiff, messageHistory string) error {
	err := s.db.Queries.UpdateTaskResult(ctx, sqlc.UpdateTaskResultParams{
		Status:         string(status),
		AgentOutput:    sql.NullString{String: agentOutput, Valid: agentOutput != ""},
		GitDiff:        sql.NullString{String: gitDiff, Valid: gitDiff != ""},
		MessageHistory: sql.NullString{String: messageHistory, Valid: messageHistory != ""},
		UpdatedAt:      sql.NullTime{Time: time.Now(), Valid: true},
		ID:             id,
	})
	if err != nil {
		return fmt.Errorf("failed to update task result: %w", err)
	}
	return nil
}

// ClearHistory clears a task's message history and agent output.
func (s *TaskService) ClearHistory(ctx context.Context, id string) error {
	err := s.db.Queries.ClearTaskHistory(ctx, sqlc.ClearTaskHistoryParams{
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:        id,
	})
	if err != nil {
		return fmt.Errorf("failed to clear task history: %w", err)
	}
	return nil
}

// ResetInProgress resets tasks stuck in in_progress back to appropriate status.
// This is called on server startup for recovery.
// - Tasks with agent_output go to review (they completed but server restarted)
// - Tasks without output go to todo (they were interrupted)
func (s *TaskService) ResetInProgress(ctx context.Context) error {
	// Tasks that have output should go to review, not todo
	result, err := s.db.Queries.ResetCompletedInProgressTasks(ctx, sqlc.ResetCompletedInProgressTasksParams{
		Status:    string(models.StatusReview),
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Status_2:  string(models.StatusInProgress),
	})
	if err != nil {
		return fmt.Errorf("failed to reset completed in-progress tasks: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows > 0 {
		slog.Info("Reset completed tasks to review", "count", rows)
	}

	// Tasks without output should go back to todo
	result, err = s.db.Queries.ResetStuckInProgressTasks(ctx, sqlc.ResetStuckInProgressTasksParams{
		Status:    string(models.StatusTodo),
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Status_2:  string(models.StatusInProgress),
	})
	if err != nil {
		return fmt.Errorf("failed to reset stuck in-progress tasks: %w", err)
	}
	rows, _ = result.RowsAffected()
	if rows > 0 {
		slog.Info("Reset stuck tasks to todo", "count", rows)
	}
	return nil
}

// AddLog adds a log entry for a task.
func (s *TaskService) AddLog(ctx context.Context, taskID, level, message string) error {
	err := s.db.Queries.CreateAgentLog(ctx, sqlc.CreateAgentLogParams{
		TaskID:  taskID,
		Level:   sql.NullString{String: level, Valid: level != ""},
		Message: message,
	})
	if err != nil {
		return fmt.Errorf("failed to add log: %w", err)
	}
	return nil
}

// GetLogs retrieves logs for a task.
func (s *TaskService) GetLogs(ctx context.Context, taskID string) ([]*models.AgentLog, error) {
	logs, err := s.db.Queries.GetAgentLogsByTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	result := make([]*models.AgentLog, len(logs))
	for i, log := range logs {
		result[i] = &models.AgentLog{
			ID:        log.ID,
			TaskID:    log.TaskID,
			Level:     models.LogLevel(log.Level.String),
			Message:   log.Message,
			CreatedAt: log.CreatedAt.Time,
		}
	}
	return result, nil
}

// sqlcTaskToModel converts a sqlc Task to a models.Task
func sqlcTaskToModel(t sqlc.Task) *models.Task {
	task := &models.Task{
		ID:        t.ID,
		ProjectID: t.ProjectID,
		Title:     t.Title,
		Intent:    t.Intent,
		Status:    models.TaskStatus(t.Status),
		Position:  int(t.Position.Int64),
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
	}
	if t.AgentOutput.Valid {
		task.AgentOutput = t.AgentOutput.String
	}
	if t.GitDiff.Valid {
		task.GitDiff = t.GitDiff.String
	}
	if t.MessageHistory.Valid {
		task.MessageHistory = t.MessageHistory.String
	}
	return task
}
