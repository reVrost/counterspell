package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/lithammer/shortuuid/v4"
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
func (s *TaskService) Create(ctx context.Context, projectID, title, intent string) (*models.Task, error) {
	id := shortuuid.New()
	now := time.Now()

	query := `
		INSERT INTO tasks (id, project_id, title, intent, status, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query, id, projectID, title, intent, models.StatusTodo, 0, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return s.Get(ctx, id)
}

// Get retrieves a task by ID.
func (s *TaskService) Get(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, project_id, title, intent, status, position, agent_output, git_diff, message_history, created_at, updated_at
		FROM tasks WHERE id = ?
	`
	var agentOutput, gitDiff, messageHistory sql.NullString
	var task models.Task
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.ProjectID, &task.Title, &task.Intent, &task.Status,
		&task.Position, &agentOutput, &gitDiff, &messageHistory, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	if agentOutput.Valid {
		task.AgentOutput = agentOutput.String
	}
	if gitDiff.Valid {
		task.GitDiff = gitDiff.String
	}
	if messageHistory.Valid {
		task.MessageHistory = messageHistory.String
	}
	return &task, nil
}

// List returns all tasks, optionally filtered by status and/or project.
func (s *TaskService) List(ctx context.Context, status *models.TaskStatus, projectID *string) ([]models.Task, error) {
	var query string
	var args []interface{}

	whereClauses := []string{}
	if status != nil {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, *status)
	}
	if projectID != nil {
		whereClauses = append(whereClauses, "project_id = ?")
		args = append(args, *projectID)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + joinStrings(whereClauses, " AND ")
	}

	query = `
		SELECT id, project_id, title, intent, status, position, agent_output, git_diff, message_history, created_at, updated_at
		FROM tasks
	` + whereClause + `
		ORDER BY status ASC, position ASC, created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Failed to close rows: %v\n", err)
		}
	}()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var agentOutput, gitDiff, messageHistory sql.NullString
		err := rows.Scan(
			&task.ID, &task.ProjectID, &task.Title, &task.Intent, &task.Status,
			&task.Position, &agentOutput, &gitDiff, &messageHistory, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		if agentOutput.Valid {
			task.AgentOutput = agentOutput.String
		}
		if gitDiff.Valid {
			task.GitDiff = gitDiff.String
		}
		if messageHistory.Valid {
			task.MessageHistory = messageHistory.String
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
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

// UpdateWithResult updates a task's status, agent output, git diff, and message history.
func (s *TaskService) UpdateWithResult(ctx context.Context, id string, status models.TaskStatus, agentOutput, gitDiff, messageHistory string) error {
	query := `
		UPDATE tasks SET status = ?, agent_output = ?, git_diff = ?, message_history = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, status, agentOutput, gitDiff, messageHistory, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update task result: %w", err)
	}
	return nil
}

// ClearHistory clears a task's message history and agent output.
func (s *TaskService) ClearHistory(ctx context.Context, id string) error {
	query := `
		UPDATE tasks SET message_history = '', agent_output = '', updated_at = ?
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, time.Now(), id)
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
	query := `
		UPDATE tasks SET status = ?, updated_at = ?
		WHERE status = ? AND (agent_output IS NOT NULL AND agent_output != '')
	`
	result, err := s.db.ExecContext(ctx, query, models.StatusReview, time.Now(), models.StatusInProgress)
	if err != nil {
		return fmt.Errorf("failed to reset completed in-progress tasks: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows > 0 {
		slog.Info("Reset completed tasks to review", "count", rows)
	}

	// Tasks without output should go back to todo
	query = `
		UPDATE tasks SET status = ?, updated_at = ?
		WHERE status = ? AND (agent_output IS NULL OR agent_output = '')
	`
	result, err = s.db.ExecContext(ctx, query, models.StatusTodo, time.Now(), models.StatusInProgress)
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
	query := `INSERT INTO agent_logs (task_id, level, message) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, taskID, level, message)
	if err != nil {
		return fmt.Errorf("failed to add log: %w", err)
	}
	return nil
}

// GetLogs retrieves logs for a task.
func (s *TaskService) GetLogs(ctx context.Context, taskID string) ([]*models.AgentLog, error) {
	query := `SELECT id, task_id, level, message, created_at FROM agent_logs WHERE task_id = ? ORDER BY id ASC`
	rows, err := s.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.AgentLog
	for rows.Next() {
		var log models.AgentLog
		if err := rows.Scan(&log.ID, &log.TaskID, &log.Level, &log.Message, &log.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, &log)
	}
	return logs, nil
}
