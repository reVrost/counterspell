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
	"github.com/revrost/counterspell/internal/db/sqlc"
	"github.com/revrost/counterspell/internal/models"
)

// Repository handles task persistence.
type Repository struct {
	Q *sqlc.Queries
}

// NewRepository creates a new task service.
func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{Q: queries}
}

func (s *Repository) GetWorkspace(ctx context.Context, workspaceID string) (sqlc.Workspace, error) {
	return s.Q.GetWorkspace(ctx, workspaceID)
}

func (s *Repository) ListWorkspaces(ctx context.Context) ([]sqlc.Workspace, error) {
	return s.Q.ListWorkspaces(ctx)
}

func (s *Repository) GetGithubConnection(ctx context.Context) (sqlc.GithubConnection, error) {
	return s.Q.GetGithubConnection(ctx)
}

func (s *Repository) GetGithubConnectionByID(ctx context.Context, githubConnectionID string) (sqlc.GithubConnection, error) {
	return s.Q.GetGithubConnectionByID(ctx, githubConnectionID)

}

func (s *Repository) CreateWorkspace(ctx context.Context, name, localPath string) (sqlc.Workspace, error) {
	now := time.Now().UnixMilli()
	workspace, err := s.Q.CreateWorkspace(ctx, sqlc.CreateWorkspaceParams{
		ID: shortuuid.New(),
		// TODO: perhaps instantiate github connection here (if exists) ?
		// GithubConnectionID: connection.ID,
		Name:      name,
		LocalPath: localPath,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return sqlc.Workspace{}, fmt.Errorf("failed to create workspace %q: %w", name, err)
	}

	return workspace, nil
}

// CreateTask creates a new task with validation.
func (s *Repository) CreateTask(ctx context.Context, workspaceID, intent string) (*models.Task, error) {
	id := shortuuid.New()
	// Validate input
	if intent == "" {
		return nil, fmt.Errorf("intent is required")
	}

	now := time.Now().UnixMilli()
	if err := s.Q.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:          id,
		WorkspaceID: sql.NullString{String: workspaceID, Valid: workspaceID != ""},
		Title:       intent, // Use intent as title for now
		Intent:      intent,
		Status:      "draft",
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		return nil, fmt.Errorf("failed to create task with id %s: %w", id, err)
	}

	return s.Get(ctx, id)
}

// CreateFromSession creates a task from a session promotion.
func (s *Repository) CreateFromSession(ctx context.Context, title, intent string) (*models.Task, error) {
	id := shortuuid.New()
	if title == "" {
		title = "Promoted session"
	}
	if intent == "" {
		intent = title
	}

	now := time.Now().UnixMilli()
	if err := s.Q.CreateTask(ctx, sqlc.CreateTaskParams{
		ID:          id,
		WorkspaceID: sql.NullString{},
		Title:       title,
		Intent:      intent,
		Status:      "draft",
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		return nil, fmt.Errorf("failed to create task with id %s: %w", id, err)
	}

	return s.Get(ctx, id)
}

// Get retrieves a task by ID.
func (s *Repository) Get(ctx context.Context, id string) (*models.Task, error) {
	task, err := s.Q.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return sqlcGetTaskRowToModel(&task), nil
}

// List retrieves all tasks.
func (s *Repository) List(ctx context.Context) ([]*models.Task, error) {
	tasks, err := s.Q.ListTasks(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Task, len(tasks))
	for i := range tasks {
		result[i] = sqlcTaskToModel(&tasks[i])
	}
	return result, nil
}

// ListWithWorkspace retrieves all tasks with workspace names.
func (s *Repository) ListWithWorkspace(ctx context.Context) ([]*models.Task, error) {
	tasks, err := s.Q.ListTasksByWorkspace(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Task, len(tasks))
	for i := range tasks {
		result[i] = sqlcTaskWithWorkspaceToModel(&tasks[i])
	}
	return result, nil
}

// ListByStatus retrieves tasks by status.
func (s *Repository) ListByStatus(ctx context.Context, status string) ([]*models.Task, error) {
	tasks, err := s.Q.ListTasksByStatus(ctx, status)
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
	validStatuses := []string{"draft", "planning", "in_progress", "review", "done", "failed"}
	if !slices.Contains(validStatuses, status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	if err := s.Q.UpdateTaskStatus(ctx, sqlc.UpdateTaskStatusParams{
		Status: status,
		ID:     id,
	}); err != nil {
		return err
	}
	return nil
}

func (s *Repository) UpdateFailedTask(ctx context.Context, id, reason string) error {
	return s.Q.UpdateFailedTask(ctx, sqlc.UpdateFailedTaskParams{
		FailedReason: sql.NullString{String: reason, Valid: reason != ""},
		ID:           id,
	})
}

// UpdateTaskTitleIntent updates a task title and intent.
func (s *Repository) UpdateTaskTitleIntent(ctx context.Context, taskID, title, intent string) error {
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}
	return s.Q.UpdateTaskTitleIntent(ctx, sqlc.UpdateTaskTitleIntentParams{
		Title:  title,
		Intent: intent,
		ID:     taskID,
	})
}

// Delete removes a task.
func (s *Repository) Delete(ctx context.Context, id string) error {
	if err := s.Q.DeleteTask(ctx, id); err != nil {
		return err
	}
	return nil
}

// GetPendingTasks retrieves all draft tasks for execution.
func (s *Repository) GetPendingTasks(ctx context.Context) ([]*models.Task, error) {
	return s.ListByStatus(ctx, "draft")
}

// GetInProgressTasks retrieves all in-progress tasks.
func (s *Repository) GetInProgressTasks(ctx context.Context) ([]*models.Task, error) {
	return s.ListByStatus(ctx, "in_progress")
}

// sqlcTaskToModel converts sqlc task to model.
func sqlcTaskToModel(task *sqlc.Task) *models.Task {
	return &models.Task{
		ID:          task.ID,
		WorkspaceID: nullableString(task.WorkspaceID),
		Title:       task.Title,
		Intent:      task.Intent,
		Status:      task.Status,
		Position:    nullableInt64(task.Position),
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

// sqlcTaskWithWorkspaceToModel converts sqlc task with workspace to model.
func sqlcTaskWithWorkspaceToModel(task *sqlc.ListTasksByWorkspaceRow) *models.Task {
	var lastMsg *string
	if msg, ok := task.LastAssistantMessage.(string); ok && msg != "" {
		copyMsg := msg
		lastMsg = &copyMsg
	}

	return &models.Task{
		ID:                   task.ID,
		WorkspaceID:          nullableString(task.WorkspaceID),
		WorkspaceName:        nullableString(task.WorkspaceName),
		Title:                task.Title,
		Intent:               task.Intent,
		Status:               task.Status,
		FailedReason:         nullableString(task.FailedReason),
		Position:             nullableInt64(task.Position),
		LastAssistantMessage: lastMsg,
		CreatedAt:            task.CreatedAt,
		UpdatedAt:            task.UpdatedAt,
	}
}

// sqlcGetTaskRowToModel converts sqlc GetTaskRow to model.
func sqlcGetTaskRowToModel(task *sqlc.GetTaskRow) *models.Task {
	return &models.Task{
		ID:            task.ID,
		WorkspaceID:   nullableString(task.WorkspaceID),
		WorkspaceName: nullableString(task.WorkspaceName),
		Title:         task.Title,
		Intent:        task.Intent,
		Status:        task.Status,
		FailedReason:  nullableString(task.FailedReason),
		Position:      nullableInt64(task.Position),
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.UpdatedAt,
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

	return s.Q.CreateMessage(ctx, sqlc.CreateMessageParams{
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

	return s.Q.CreateMessage(ctx, sqlc.CreateMessageParams{
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
	return s.Q.GetMessagesByTask(ctx, taskID)
}

// --- Agent Run Operations ---

// GetTaskWithDetails retrieves a task with all related data for TaskResponse.
// This uses multiple efficient queries to get the task, all messages, artifacts, and agent runs.
// GetTaskWithDetails retrieves a task with all related data for TaskResponse.
// This uses sqlc queries to get task, messages, artifacts, and agent runs.
func (s *Repository) GetTaskWithDetails(ctx context.Context, taskID string) (*models.TaskResponse, error) {
	// Get the base task
	task, err := s.Q.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Get all messages for the task
	messages, err := s.Q.GetMessagesByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Get all artifacts for the task
	artifacts, err := s.Q.GetArtifactsByTask(ctx, taskID)
	if err != nil {
		return nil, err
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
	}, nil
}

// CreateAgentRun creates a new agent run.
func (s *Repository) CreateAgentRun(ctx context.Context, taskID, prompt, agentBackend, provider, model string) (string, error) {
	id := shortuuid.New()
	now := time.Now().UnixMilli()

	if err := s.Q.CreateAgentRun(ctx, sqlc.CreateAgentRunParams{
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
	return s.Q.UpdateAgentRunCompleted(ctx, sqlc.UpdateAgentRunCompletedParams{
		CompletedAt: sql.NullTime{Time: now, Valid: true},
		ID:          runID,
	})
}

// UpdateAgentRunBackendSessionID saves the backend's session ID.
func (s *Repository) UpdateAgentRunBackendSessionID(ctx context.Context, runID, sessionID string) error {
	return s.Q.UpdateAgentRunBackendSessionID(ctx, sqlc.UpdateAgentRunBackendSessionIDParams{
		BackendSessionID: sql.NullString{String: sessionID, Valid: sessionID != ""},
		ID:               runID,
	})
}

// GetLatestAgentRun retrieves the most recent agent run for a task.
func (s *Repository) GetLatestAgentRun(ctx context.Context, taskID string) (*sqlc.AgentRun, error) {
	run, err := s.Q.GetLatestRun(ctx, taskID)
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
	run, err := s.Q.GetAgentRun(ctx, runID)
	if err != nil {
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

// --- Session Operations ---

// CreateSession inserts a new session.
func (s *Repository) CreateSession(ctx context.Context, session *models.Session) (*models.Session, error) {
	if session == nil {
		return nil, fmt.Errorf("session is required")
	}

	if err := s.Q.CreateSession(ctx, sqlc.CreateSessionParams{
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
	row, err := s.Q.GetSession(ctx, sessionID)
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
	row, err := s.Q.GetSessionByBackendExternal(ctx, sqlc.GetSessionByBackendExternalParams{
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
	rows, err := s.Q.ListSessions(ctx)
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
	return s.Q.UpdateSession(ctx, sqlc.UpdateSessionParams{
		BackendSessionID: sql.NullString{String: backendSessionID, Valid: backendSessionID != ""},
		Title:            sql.NullString{String: title, Valid: title != ""},
		LastMessageAt:    sql.NullInt64{Int64: valueOrZero(lastMessageAt), Valid: lastMessageAt != nil},
		UpdatedAt:        time.Now().UnixMilli(),
		ID:               sessionID,
	})
}

// UpdateSessionBackendSessionID updates session backend session id.
func (s *Repository) UpdateSessionBackendSessionID(ctx context.Context, sessionID, backendSessionID string) error {
	return s.Q.UpdateSessionBackendSessionID(ctx, sqlc.UpdateSessionBackendSessionIDParams{
		BackendSessionID: sql.NullString{String: backendSessionID, Valid: backendSessionID != ""},
		UpdatedAt:        time.Now().UnixMilli(),
		ID:               sessionID,
	})
}

// UpdateSessionTitle updates session title.
func (s *Repository) UpdateSessionTitle(ctx context.Context, sessionID, title string) error {
	return s.Q.UpdateSessionTitle(ctx, sqlc.UpdateSessionTitleParams{
		Title:     sql.NullString{String: title, Valid: title != ""},
		UpdatedAt: time.Now().UnixMilli(),
		ID:        sessionID,
	})
}

// GetSessionNextSequence returns the next sequence number for a session message.
func (s *Repository) GetSessionNextSequence(ctx context.Context, sessionID string) (int64, error) {
	next, err := s.Q.GetSessionNextSequence(ctx, sessionID)
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
	return s.Q.CreateSessionMessage(ctx, sqlc.CreateSessionMessageParams{
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
	rows, err := s.Q.ListSessionMessages(ctx, sessionID)
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
