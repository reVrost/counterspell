package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
)

// ChatMessage represents a chat message (renamed to avoid collision).
type ChatMessage struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	RunID     string    `json:"run_id"`
	Role      string    `json:"role"` // "system", "user", "assistant", "tool"
	Content   string    `json:"content"`
	ToolID    string    `json:"tool_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// MessageService manages message history.
type MessageService struct {
	db *db.DB
}

// NewMessageService creates a new message service.
func NewMessageService(database *db.DB) *MessageService {
	return &MessageService{db: database}
}

// CreateMessage creates a new message.
func (s *MessageService) CreateMessage(ctx context.Context, taskID, runID, role, content, toolID string) (*ChatMessage, error) {
	id := uuid.New().String()
	now := time.Now()

	runIDParam := sql.NullString{String: runID, Valid: runID != ""}
	toolIDParam := sql.NullString{String: toolID, Valid: toolID != ""}
	createdAtParam := sql.NullTime{Time: now, Valid: true}

	if err := s.db.Queries.CreateMessage(ctx, sqlc.CreateMessageParams{
		ID:        id,
		TaskID:    taskID,
		RunID:     runIDParam,
		Role:      role,
		Content:   content,
		ToolID:    toolIDParam,
		CreatedAt: createdAtParam.Time.UnixMilli(),
	}); err != nil {
		return nil, err
	}

	return s.GetMessage(ctx, id)
}

// GetMessage retrieves a message by ID.
func (s *MessageService) GetMessage(ctx context.Context, id string) (*ChatMessage, error) {
	msg, err := s.db.Queries.GetMessage(ctx, id)
	if err != nil {
		return nil, err
	}
	return sqlcMessageToModel(&msg), nil
}

// GetMessagesByTask retrieves all messages for a task.
func (s *MessageService) GetMessagesByTask(ctx context.Context, taskID string) ([]*ChatMessage, error) {
	messages, err := s.db.Queries.GetMessagesByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	result := make([]*ChatMessage, len(messages))
	for i := range messages {
		result[i] = sqlcMessageToModel(&messages[i])
	}
	return result, nil
}

// GetMessagesByRun retrieves all messages for a specific agent run.
func (s *MessageService) GetMessagesByRun(ctx context.Context, runID string) ([]*ChatMessage, error) {
	runIDParam := sql.NullString{String: runID, Valid: runID != ""}
	messages, err := s.db.Queries.GetMessagesByRun(ctx, runIDParam)
	if err != nil {
		return nil, err
	}

	result := make([]*ChatMessage, len(messages))
	for i := range messages {
		result[i] = sqlcMessageToModel(&messages[i])
	}
	return result, nil
}

// GetRecentMessages retrieves recent messages for a task (for context).
func (s *MessageService) GetRecentMessages(ctx context.Context, taskID string, limit int) ([]*ChatMessage, error) {
	messages, err := s.db.Queries.GetRecentMessages(ctx, sqlc.GetRecentMessagesParams{
		TaskID: taskID,
		Limit:  int64(limit),
	})
	if err != nil {
		return nil, err
	}

	result := make([]*ChatMessage, len(messages))
	for i := range messages {
		result[i] = sqlcMessageToModel(&messages[i])
	}
	return result, nil
}

// CreateSystemMessage creates a system message for a task.
func (s *MessageService) CreateSystemMessage(ctx context.Context, taskID, content string) error {
	_, err := s.CreateMessage(ctx, taskID, "", "system", content, "")
	return err
}

// CreateUserMessage creates a user message.
func (s *MessageService) CreateUserMessage(ctx context.Context, taskID, runID, content string) error {
	_, err := s.CreateMessage(ctx, taskID, runID, "user", content, "")
	return err
}

// CreateAssistantMessage creates an assistant message.
func (s *MessageService) CreateAssistantMessage(ctx context.Context, taskID, runID, content string) error {
	_, err := s.CreateMessage(ctx, taskID, runID, "assistant", content, "")
	return err
}

// CreateToolMessage creates a tool response message.
func (s *MessageService) CreateToolMessage(ctx context.Context, taskID, runID, toolID, content string) error {
	_, err := s.CreateMessage(ctx, taskID, runID, "tool", content, toolID)
	return err
}

// BuildContextForTask builds message context for AI from task history.
func (s *MessageService) BuildContextForTask(ctx context.Context, taskID string) ([]Message, error) {
	// Get recent messages (last 50)
	messages, err := s.GetRecentMessages(ctx, taskID, 50)
	if err != nil {
		return nil, err
	}

	// Convert to agent message format
	result := make([]Message, len(messages))
	for i, msg := range messages {
		result[i] = Message{
			Role:    msg.Role,
			Content: msg.Content,
			ToolID:  msg.ToolID,
		}
	}

	return result, nil
}

// sqlcMessageToModel converts sqlc message to model.
func sqlcMessageToModel(msg *sqlc.Message) *ChatMessage {
	createdAt := time.UnixMilli(msg.CreatedAt)
	return &ChatMessage{
		ID:        msg.ID,
		TaskID:    msg.TaskID,
		RunID:     msg.RunID.String,
		Role:      msg.Role,
		Content:   msg.Content,
		ToolID:    msg.ToolID.String,
		CreatedAt: createdAt,
	}
}
