package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"github.com/revrost/counterspell/internal/agent"
	"github.com/revrost/counterspell/internal/db"
	"github.com/revrost/counterspell/internal/db/sqlc"
	"github.com/revrost/counterspell/internal/models"
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
		ID:               id,
		RepositoryID:     sql.NullString{String: repositoryID, Valid: repositoryID != ""},
		SessionID:        sql.NullString{},
		Title:            intent, // Use intent as title for now
		Intent:           intent,
		PromotedSnapshot: sql.NullString{},
		Status:           "pending",
		CreatedAt:        now,
		UpdatedAt:        now,
	}); err != nil {
		return nil, fmt.Errorf("failed to create task with id %s: %w", id, err)
	}

	return s.Get(ctx, id)
}

// CreateFromSession creates a task from a session promotion.
func (s *Repository) CreateFromSession(ctx context.Context, sessionID, title, intent, snapshot string) (*models.Task, error) {
	id := shortuuid.New()
	if title == "" {
		title = "Promoted session"
	}
	if intent == "" {
		intent = title
	}

	now := time.Now().UnixMilli()
	if err := s.db.Queries.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:               id,
		RepositoryID:     sql.NullString{},
		SessionID:        sql.NullString{String: sessionID, Valid: sessionID != ""},
		Title:            title,
		Intent:           intent,
		PromotedSnapshot: sql.NullString{String: snapshot, Valid: snapshot != ""},
		Status:           "pending",
		CreatedAt:        now,
		UpdatedAt:        now,
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
	return sqlcGetTaskRowToModel(&task), nil
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

// GetTaskBySessionID retrieves a task by session ID.
func (s *Repository) GetTaskBySessionID(ctx context.Context, sessionID string) (*models.Task, error) {
	if sessionID == "" {
		return nil, nil
	}
	task, err := s.db.Queries.GetTaskBySessionID(ctx, sql.NullString{String: sessionID, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return sqlcTaskToModel(&task), nil
}

// UpdateTaskTitleIntent updates a task title and intent.
func (s *Repository) UpdateTaskTitleIntent(ctx context.Context, taskID, title, intent string) error {
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}
	return s.db.Queries.UpdateTaskTitleIntent(ctx, sqlc.UpdateTaskTitleIntentParams{
		Title:  title,
		Intent: intent,
		ID:     taskID,
	})
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
		ID:               task.ID,
		RepositoryID:     nullableString(task.RepositoryID),
		SessionID:        nullableString(task.SessionID),
		Title:            task.Title,
		Intent:           task.Intent,
		PromotedSnapshot: nullableString(task.PromotedSnapshot),
		Status:           task.Status,
		Position:         nullableInt64(task.Position),
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
	}
}

// sqlcTaskWithRepoToModel converts sqlc task with repository to model.
func sqlcTaskWithRepoToModel(task *sqlc.ListTasksWithRepositoryRow) *models.Task {
	var lastMsg *string
	if msg, ok := task.LastAssistantMessage.(string); ok && msg != "" {
		copyMsg := msg
		lastMsg = &copyMsg
	}

	return &models.Task{
		ID:                   task.ID,
		RepositoryID:         nullableString(task.RepositoryID),
		RepositoryName:       nullableString(task.RepositoryName),
		SessionID:            nullableString(task.SessionID),
		Title:                task.Title,
		Intent:               task.Intent,
		PromotedSnapshot:     nullableString(task.PromotedSnapshot),
		Status:               task.Status,
		Position:             nullableInt64(task.Position),
		LastAssistantMessage: lastMsg,
		CreatedAt:            task.CreatedAt,
		UpdatedAt:            task.UpdatedAt,
	}
}

// sqlcGetTaskRowToModel converts sqlc GetTaskRow to model.
func sqlcGetTaskRowToModel(task *sqlc.GetTaskRow) *models.Task {
	return &models.Task{
		ID:               task.ID,
		RepositoryID:     nullableString(task.RepositoryID),
		RepositoryName:   nullableString(task.RepositoryName),
		SessionID:        nullableString(task.SessionID),
		Title:            task.Title,
		Intent:           task.Intent,
		PromotedSnapshot: nullableString(task.PromotedSnapshot),
		Status:           task.Status,
		Position:         nullableInt64(task.Position),
		CreatedAt:        task.CreatedAt,
		UpdatedAt:        task.UpdatedAt,
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
func (s *Repository) CreateMessage(ctx context.Context, taskID, runID, role, content string) error {
	id := shortuuid.New()
	now := time.Now().UnixMilli()

	return s.db.Queries.CreateMessage(ctx, sqlc.CreateMessageParams{
		ID:      id,
		TaskID:  taskID,
		RunID:   runID,
		Role:    role,
		Content: content,
		Parts:   "[]",
		// Model:     sql.NullString{String: model, Valid: model != ""},
		// Provider:  sql.NullString{String: provider, Valid: provider != ""},
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// CreateMessageWithParts creates a new message with structured parts.
func (s *Repository) CreateMessageWithParts(ctx context.Context, taskID, runID, role, content, parts string) error {
	id := shortuuid.New()
	now := time.Now().UnixMilli()
	if strings.TrimSpace(parts) == "" {
		parts = "[]"
	}

	return s.db.Queries.CreateMessage(ctx, sqlc.CreateMessageParams{
		ID:        id,
		TaskID:    taskID,
		RunID:     runID,
		Role:      role,
		Content:   content,
		Parts:     parts,
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
// This uses multiple efficient queries to get the task, all messages, artifacts, and agent runs.
// GetTaskWithDetails retrieves a task with all related data for TaskResponse.
// This uses sqlc queries to get task, messages, artifacts, and agent runs.
func (s *Repository) GetTaskWithDetails(ctx context.Context, taskID string) (*models.TaskResponse, error) {
	// Get the base task
	task, err := s.db.Queries.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Get all messages for the task
	messages, err := s.db.Queries.GetMessagesByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Get all artifacts for the task
	artifacts, err := s.db.Queries.GetArtifactsByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Get all agent runs for the task
	agentRuns, err := s.db.Queries.ListAgentRunsByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Build agent runs with nested messages and artifacts
	agentRunsWithDetails := make([]models.AgentRunWithDetails, len(agentRuns))
	for i, ar := range agentRuns {
		// Find messages for this run
		runMessages := make([]models.Message, 0)
		for _, msg := range messages {
			if msg.RunID == ar.ID {
				runMessages = append(runMessages, models.Message{
					ID:         msg.ID,
					TaskID:     msg.TaskID,
					RunID:      &msg.RunID,
					Role:       msg.Role,
					Parts:      msg.Parts,
					Model:      nullableString(msg.Model),
					Provider:   nullableString(msg.Provider),
					Content:    msg.Content,
					ToolID:     nullableString(msg.ToolID),
					CreatedAt:  msg.CreatedAt,
					UpdatedAt:  msg.UpdatedAt,
					FinishedAt: nullableInt64(msg.FinishedAt),
				})
			}
		}

		// Find artifacts for this run
		runArtifacts := make([]models.Artifact, 0)
		for _, art := range artifacts {
			if art.RunID == ar.ID {
				runArtifacts = append(runArtifacts, models.Artifact{
					ID:        art.ID,
					RunID:     art.RunID,
					Path:      art.Path,
					Content:   art.Content,
					Version:   art.Version,
					CreatedAt: art.CreatedAt,
					UpdatedAt: art.UpdatedAt,
				})
			}
		}

		agentRunsWithDetails[i] = models.AgentRunWithDetails{
			ID:               ar.ID,
			TaskID:           ar.TaskID,
			Prompt:           ar.Prompt,
			AgentBackend:     ar.AgentBackend,
			SummaryMessageID: nullableString(ar.SummaryMessageID),
			Cost:             ar.Cost,
			MessageCount:     ar.MessageCount,
			PromptTokens:     ar.PromptTokens,
			CompletionTokens: ar.CompletionTokens,
			CompletedAt:      nullableInt64FromTime(ar.CompletedAt),
			CreatedAt:        ar.CreatedAt,
			UpdatedAt:        ar.UpdatedAt,
			Messages:         runMessages,
			Artifacts:        runArtifacts,
		}
	}

	// Top-level messages should include ALL messages for the task
	taskMessages := make([]models.Message, len(messages))
	for i, msg := range messages {
		taskMessages[i] = models.Message{
			ID:         msg.ID,
			TaskID:     msg.TaskID,
			RunID:      &msg.RunID,
			Role:       msg.Role,
			Parts:      msg.Parts,
			Model:      nullableString(msg.Model),
			Provider:   nullableString(msg.Provider),
			Content:    msg.Content,
			ToolID:     nullableString(msg.ToolID),
			CreatedAt:  msg.CreatedAt,
			UpdatedAt:  msg.UpdatedAt,
			FinishedAt: nullableInt64(msg.FinishedAt),
		}
	}

	// Convert artifacts to models
	taskArtifacts := make([]models.Artifact, len(artifacts))
	for i, art := range artifacts {
		taskArtifacts[i] = models.Artifact{
			ID:        art.ID,
			RunID:     art.RunID,
			Path:      art.Path,
			Content:   art.Content,
			Version:   art.Version,
			CreatedAt: art.CreatedAt,
			UpdatedAt: art.UpdatedAt,
		}
	}

	return &models.TaskResponse{
		Task:      *sqlcGetTaskRowToModel(&task),
		Messages:  taskMessages,
		Artifacts: taskArtifacts,
		AgentRuns: agentRunsWithDetails,
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

// UpdateAgentRunBackendSessionID saves the backend's session ID.
func (s *Repository) UpdateAgentRunBackendSessionID(ctx context.Context, runID, sessionID string) error {
	return s.db.Queries.UpdateAgentRunBackendSessionID(ctx, sqlc.UpdateAgentRunBackendSessionIDParams{
		BackendSessionID: sql.NullString{String: sessionID, Valid: sessionID != ""},
		ID:               runID,
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

// GetAgentRun retrieves an agent run by ID.
func (s *Repository) GetAgentRun(ctx context.Context, runID string) (*sqlc.AgentRun, error) {
	run, err := s.db.Queries.GetAgentRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// GetAgentRunByBackendSessionID retrieves an agent run by backend session ID.
func (s *Repository) GetAgentRunByBackendSessionID(ctx context.Context, backend, sessionID string) (*sqlc.AgentRun, error) {
	if backend == "" || sessionID == "" {
		return nil, nil
	}

	row := s.db.DB.QueryRowContext(ctx, `SELECT id, task_id, prompt, agent_backend, provider, model, summary_message_id, backend_session_id, cost, message_count, prompt_tokens, completion_tokens, completed_at, created_at, updated_at FROM agent_runs WHERE agent_backend = ? AND backend_session_id = ? ORDER BY created_at DESC LIMIT 1`, backend, sessionID)
	var run sqlc.AgentRun
	err := row.Scan(
		&run.ID,
		&run.TaskID,
		&run.Prompt,
		&run.AgentBackend,
		&run.Provider,
		&run.Model,
		&run.SummaryMessageID,
		&run.BackendSessionID,
		&run.Cost,
		&run.MessageCount,
		&run.PromptTokens,
		&run.CompletionTokens,
		&run.CompletedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &run, nil
}

// ConvertMessagesToJSON converts sqlc.Message to JSON format for agent state restoration.
func ConvertMessagesToJSON(messages []sqlc.Message) (string, error) {
	type Message struct {
		Role    string               `json:"role"`
		Content []agent.ContentBlock `json:"content"`
	}

	result := make([]Message, 0, len(messages))
	for _, msg := range messages {
		var blocks []agent.ContentBlock
		if strings.TrimSpace(msg.Parts) != "" && strings.TrimSpace(msg.Parts) != "[]" {
			_ = json.Unmarshal([]byte(msg.Parts), &blocks)
		}
		if len(blocks) == 0 {
			if strings.TrimSpace(msg.Content) == "" {
				continue
			}
			blocks = []agent.ContentBlock{{Type: "text", Text: msg.Content}}
		}

		filtered := blocks[:0]
		for _, block := range blocks {
			if block.Type == "thinking" {
				continue
			}
			filtered = append(filtered, block)
		}

		role := msg.Role
		if len(blocks) > 0 {
			hasToolResult := false
			hasToolUse := false
			for _, block := range blocks {
				if block.Type == "tool_result" {
					hasToolResult = true
				} else if block.Type == "tool_use" {
					hasToolUse = true
				}
			}
			if hasToolResult {
				role = "user"
			} else if hasToolUse {
				role = "assistant"
			}
		} else {
			if role == "tool_result" {
				role = "user"
			} else if role == "tool" {
				role = "assistant"
			}
		}

		result = append(result, Message{
			Role:    role,
			Content: filtered,
		})
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal messages: %w", err)
	}
	return string(jsonData), nil
}

// nullableInt64FromTime converts sql.NullTime to *int64.
func nullableInt64FromTime(t sql.NullTime) *int64 {
	if t.Valid {
		unixMillis := t.Time.UnixMilli()
		return &unixMillis
	}
	return nil
}

// --- Session Operations ---

// CreateSession inserts a new session.
func (s *Repository) CreateSession(ctx context.Context, session *models.Session) (*models.Session, error) {
	if session == nil {
		return nil, fmt.Errorf("session is required")
	}

	if err := s.db.Queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:               session.ID,
		AgentBackend:     session.AgentBackend,
		ExternalID:       sql.NullString{String: valueOrEmpty(session.ExternalID), Valid: session.ExternalID != nil},
		BackendSessionID: sql.NullString{String: valueOrEmpty(session.BackendSessionID), Valid: session.BackendSessionID != nil},
		Title:            sql.NullString{String: valueOrEmpty(session.Title), Valid: session.Title != nil},
		MessageCount:     session.MessageCount,
		LastMessageAt:    sql.NullInt64{Int64: valueOrZero(session.LastMessageAt), Valid: session.LastMessageAt != nil},
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
	}); err != nil {
		return nil, err
	}

	return s.GetSession(ctx, session.ID)
}

// GetSession retrieves a session by ID.
func (s *Repository) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	row, err := s.db.Queries.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return sqlcSessionToModel(&row), nil
}

// GetSessionByBackendExternal retrieves a session by backend/external ID.
func (s *Repository) GetSessionByBackendExternal(ctx context.Context, backend, externalID string) (*models.Session, error) {
	if backend == "" || externalID == "" {
		return nil, nil
	}
	row, err := s.db.Queries.GetSessionByBackendExternal(ctx, sqlc.GetSessionByBackendExternalParams{
		AgentBackend: backend,
		ExternalID:   sql.NullString{String: externalID, Valid: externalID != ""},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return sqlcSessionToModel(&row), nil
}

// ListSessions retrieves all sessions.
func (s *Repository) ListSessions(ctx context.Context) ([]*models.Session, error) {
	rows, err := s.db.Queries.ListSessions(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*models.Session, len(rows))
	for i := range rows {
		result[i] = sqlcSessionToModel(&rows[i])
	}
	return result, nil
}

// UpdateSession updates session metadata.
func (s *Repository) UpdateSession(ctx context.Context, sessionID, backendSessionID, title string, lastMessageAt *int64) error {
	return s.db.Queries.UpdateSession(ctx, sqlc.UpdateSessionParams{
		BackendSessionID: sql.NullString{String: backendSessionID, Valid: backendSessionID != ""},
		Title:            sql.NullString{String: title, Valid: title != ""},
		LastMessageAt:    sql.NullInt64{Int64: valueOrZero(lastMessageAt), Valid: lastMessageAt != nil},
		UpdatedAt:        time.Now().UnixMilli(),
		ID:               sessionID,
	})
}

// UpdateSessionBackendSessionID updates session backend session id.
func (s *Repository) UpdateSessionBackendSessionID(ctx context.Context, sessionID, backendSessionID string) error {
	return s.db.Queries.UpdateSessionBackendSessionID(ctx, sqlc.UpdateSessionBackendSessionIDParams{
		BackendSessionID: sql.NullString{String: backendSessionID, Valid: backendSessionID != ""},
		UpdatedAt:        time.Now().UnixMilli(),
		ID:               sessionID,
	})
}

// UpdateSessionTitle updates session title.
func (s *Repository) UpdateSessionTitle(ctx context.Context, sessionID, title string) error {
	return s.db.Queries.UpdateSessionTitle(ctx, sqlc.UpdateSessionTitleParams{
		Title:     sql.NullString{String: title, Valid: title != ""},
		UpdatedAt: time.Now().UnixMilli(),
		ID:        sessionID,
	})
}

// GetSessionNextSequence returns the next sequence number for a session message.
func (s *Repository) GetSessionNextSequence(ctx context.Context, sessionID string) (int64, error) {
	next, err := s.db.Queries.GetSessionNextSequence(ctx, sessionID)
	if err != nil {
		return 0, err
	}
	return next, nil
}

// CreateSessionMessage inserts a session message.
func (s *Repository) CreateSessionMessage(
	ctx context.Context,
	sessionID string,
	sequence int64,
	role string,
	kind string,
	content string,
	toolName string,
	toolCallID string,
	rawJSON string,
	createdAt int64,
) error {
	id := shortuuid.New()
	return s.db.Queries.CreateSessionMessage(ctx, sqlc.CreateSessionMessageParams{
		ID:         id,
		SessionID:  sessionID,
		Sequence:   sequence,
		Role:       role,
		Kind:       kind,
		Content:    sql.NullString{String: content, Valid: content != ""},
		ToolName:   sql.NullString{String: toolName, Valid: toolName != ""},
		ToolCallID: sql.NullString{String: toolCallID, Valid: toolCallID != ""},
		RawJson:    rawJSON,
		CreatedAt:  createdAt,
	})
}

// ListSessionMessages retrieves messages for a session.
func (s *Repository) ListSessionMessages(ctx context.Context, sessionID string) ([]models.SessionMessage, error) {
	rows, err := s.db.Queries.ListSessionMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	result := make([]models.SessionMessage, len(rows))
	for i := range rows {
		result[i] = sqlcSessionMessageToModel(&rows[i])
	}
	return result, nil
}

func sqlcSessionToModel(session *sqlc.Session) *models.Session {
	return &models.Session{
		ID:               session.ID,
		AgentBackend:     session.AgentBackend,
		ExternalID:       nullableString(session.ExternalID),
		BackendSessionID: nullableString(session.BackendSessionID),
		Title:            nullableString(session.Title),
		MessageCount:     session.MessageCount,
		LastMessageAt:    nullableInt64(session.LastMessageAt),
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
	}
}

func sqlcSessionMessageToModel(msg *sqlc.SessionMessage) models.SessionMessage {
	return models.SessionMessage{
		ID:         msg.ID,
		SessionID:  msg.SessionID,
		Sequence:   msg.Sequence,
		Role:       msg.Role,
		Kind:       msg.Kind,
		Content:    nullableString(msg.Content),
		ToolName:   nullableString(msg.ToolName),
		ToolCallID: nullableString(msg.ToolCallID),
		RawJSON:    msg.RawJson,
		CreatedAt:  msg.CreatedAt,
	}
}

func valueOrEmpty(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

func valueOrZero(val *int64) int64 {
	if val == nil {
		return 0
	}
	return *val
}
