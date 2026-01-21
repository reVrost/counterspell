package services

import (
	"context"
	"database/sql"
	"encoding/json"
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
		return nil, fmt.Errorf("failed to create task with id %s: %w", id, err)
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

// ListWithRepository retrieves all tasks with repository names.
func (s *Repository) ListWithRepository(ctx context.Context) ([]*models.Task, error) {
	tasks, err := s.db.Queries.ListTasksWithRepository(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Task, len(tasks))
	for i := range tasks {
		result[i] = sqlcTaskWithRepoToModel(&tasks[i])
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
	validStatuses := []string{"pending", "planning", "in_progress", "review", "done", "failed"}
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

// sqlcTaskWithRepoToModel converts sqlc task with repository to model.
func sqlcTaskWithRepoToModel(task *sqlc.ListTasksWithRepositoryRow) *models.Task {
	return &models.Task{
		ID:             task.ID,
		RepositoryID:   nullableString(task.RepositoryID),
		RepositoryName: nullableString(task.RepositoryName),
		Title:          task.Title,
		Intent:         task.Intent,
		Status:         task.Status,
		Position:       nullableInt64(task.Position),
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
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

// --- Message Operations ---

// CreateMessage creates a new message.
func (s *Repository) CreateMessage(ctx context.Context, taskID, runID, role, content, model, provider string) error {
	id := shortuuid.New()
	now := time.Now().UnixMilli()

	return s.db.Queries.CreateMessage(ctx, sqlc.CreateMessageParams{
		ID:        id,
		TaskID:    taskID,
		RunID:     sql.NullString{String: runID, Valid: runID != ""},
		Role:      role,
		Content:   content,
		Parts:     "[]",
		Model:     sql.NullString{String: model, Valid: model != ""},
		Provider:  sql.NullString{String: provider, Valid: provider != ""},
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// GetMessagesByTask retrieves all messages for a task.
func (s *Repository) GetMessagesByTask(ctx context.Context, taskID string) ([]sqlc.Message, error) {
	return s.db.Queries.GetMessagesByTask(ctx, taskID)
}

// --- Agent Run Operations ---

// GetTaskWithDetails retrieves a task with all related data for TaskResponse.
// This includes the task, all messages, artifacts, latest agent run, and message count.
func (s *Repository) GetTaskWithDetails(ctx context.Context, taskID string) (*models.TaskResponse, error) {
	// Get the base task
	task, err := s.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Get all messages for the task
	messages, err := s.db.Queries.GetMessagesByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Convert sqlc messages to models
	taskMessages := make([]models.Message, len(messages))
	for i, msg := range messages {
		taskMessages[i] = sqlcMessageToModel(&msg)
	}

	// Get all artifacts for the task
	artifacts, err := s.db.Queries.GetArtifactsByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Convert sqlc artifacts to models
	taskArtifacts := make([]models.Artifact, len(artifacts))
	for i, art := range artifacts {
		taskArtifacts[i] = sqlcArtifactToModel(&art)
	}

	// Get the latest agent run
	latestRun, err := s.GetLatestAgentRun(ctx, taskID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var latestAgentRun *models.AgentRun
	var messageCount int64
	if latestRun != nil {
		latestAgentRun = sqlcAgentRunToModel(latestRun)
		messageCount = latestRun.MessageCount
	}

	// If no runs, count the messages directly
	if messageCount == 0 {
		messageCount = int64(len(messages))
	}

	return &models.TaskResponse{
		Task:          *task,
		Messages:      taskMessages,
		Artifacts:     taskArtifacts,
		LatestAgentRun: latestAgentRun,
		MessageCount:  messageCount,
	}, nil
}

// CreateAgentRun creates a new agent run.
func (s *Repository) CreateAgentRun(ctx context.Context, taskID, prompt, agentBackend, provider, model string) (string, error) {
	id := shortuuid.New()
	now := time.Now().UnixMilli()

	if err := s.db.Queries.CreateAgentRun(ctx, sqlc.CreateAgentRunParams{
		ID:           id,
		TaskID:       taskID,
		Prompt:       prompt,
		AgentBackend: agentBackend,
		Provider:     sql.NullString{String: provider, Valid: provider != ""},
		Model:        sql.NullString{String: model, Valid: model != ""},
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		return "", err
	}

	return id, nil
}

// UpdateAgentRunCompleted marks an agent run as completed.
func (s *Repository) UpdateAgentRunCompleted(ctx context.Context, runID string) error {
	now := time.Now()
	return s.db.Queries.UpdateAgentRunCompleted(ctx, sqlc.UpdateAgentRunCompletedParams{
		CompletedAt: sql.NullTime{Time: now, Valid: true},
		ID:          runID,
	})
}

// GetLatestAgentRun retrieves the most recent agent run for a task.
func (s *Repository) GetLatestAgentRun(ctx context.Context, taskID string) (*sqlc.AgentRun, error) {
	run, err := s.db.Queries.GetLatestRun(ctx, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &run, nil
}

// ConvertMessagesToJSON converts sqlc.Message to JSON format for agent state restoration.
func ConvertMessagesToJSON(messages []sqlc.Message) (string, error) {
	type ContentBlock struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	}
	type Message struct {
		Role    string         `json:"role"`
		Content []ContentBlock `json:"content"`
	}

	result := make([]Message, 0, len(messages))
	for _, msg := range messages {
		result = append(result, Message{
			Role: msg.Role,
			Content: []ContentBlock{
				{Type: "text", Text: msg.Content},
			},
		})
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal messages: %w", err)
	}
	return string(jsonData), nil
}

// sqlcMessageToModel converts sqlc.Message to models.Message.
func sqlcMessageToModel(msg *sqlc.Message) models.Message {
	return models.Message{
		ID:        msg.ID,
		TaskID:    msg.TaskID,
		RunID:     nullableString(msg.RunID),
		Role:      msg.Role,
		Parts:     msg.Parts,
		Model:     nullableString(msg.Model),
		Provider:  nullableString(msg.Provider),
		Content:   msg.Content,
		ToolID:    nullableString(msg.ToolID),
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
		FinishedAt: nullableInt64(msg.FinishedAt),
	}
}

// sqlcArtifactToModel converts sqlc.Artifact to models.Artifact.
func sqlcArtifactToModel(art *sqlc.Artifact) models.Artifact {
	return models.Artifact{
		ID:        art.ID,
		RunID:     art.RunID,
		Path:      art.Path,
		Content:   art.Content,
		Version:   art.Version,
		CreatedAt: art.CreatedAt,
		UpdatedAt: art.UpdatedAt,
	}
}

// sqlcAgentRunToModel converts sqlc.AgentRun to models.AgentRun.
func sqlcAgentRunToModel(run *sqlc.AgentRun) *models.AgentRun {
	return &models.AgentRun{
		ID:               run.ID,
		TaskID:           run.TaskID,
		Prompt:           run.Prompt,
		AgentBackend:     run.AgentBackend,
		SummaryMessageID: nullableString(run.SummaryMessageID),
		Cost:             run.Cost,
		MessageCount:     run.MessageCount,
		PromptTokens:     run.PromptTokens,
		CompletionTokens: run.CompletionTokens,
		CompletedAt:      nullableInt64FromTime(run.CompletedAt),
		CreatedAt:        run.CreatedAt,
		UpdatedAt:        run.UpdatedAt,
	}
}

// nullableInt64FromTime converts sql.NullTime to *int64.
func nullableInt64FromTime(t sql.NullTime) *int64 {
	if t.Valid {
		unixMillis := t.Time.UnixMilli()
		return &unixMillis
	}
	return nil
}
