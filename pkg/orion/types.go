package orion

import (
	"context"
	"fmt"

	"charm.land/fantasy"
)

// Agent represents the core agentic interface for LLM-driven task execution.
// It manages sessions, tools, and the orchestration loop that enables
// multi-turn conversations with tool execution.
type Agent interface {
	// Run executes an agent call with streaming support.
	// Returns the final result or an error.
	Run(context.Context, AgentCall) (*fantasy.AgentResult, error)

	// SetModels updates the large and small models used by the agent.
	// The large model is used for main tasks, the small model for
	// lightweight operations like title generation.
	SetModels(large fantasy.LanguageModel, small fantasy.LanguageModel)

	// SetTools registers available tools with the agent.
	// Tools are discovered and used by the LLM during execution.
	SetTools(tools []fantasy.AgentTool)

	// Cancel cancels an active request for the given session ID.
	Cancel(sessionID string)

	// CancelAll cancels all active requests across all sessions.
	CancelAll()

	// IsSessionBusy checks if a session is currently processing a request.
	IsSessionBusy(sessionID string) bool

	// IsBusy checks if any session is currently processing a request.
	IsBusy() bool

	// QueuedPrompts returns the number of queued prompts for a session.
	QueuedPrompts(sessionID string) int

	// QueuedPromptsList returns the list of queued prompts for a session.
	QueuedPromptsList(sessionID string) []string

	// ClearQueue clears all queued prompts for a session.
	ClearQueue(sessionID string)

	// Summarize creates a summary of the session conversation.
	// Useful for reducing context window usage in long conversations.
	Summarize(context.Context, string, fantasy.ProviderOptions) error

	// Model returns the currently configured large model.
	Model() fantasy.LanguageModel
}

// SessionService defines the interface for session persistence.
// Sessions manage state, metadata, and usage tracking for conversations.
type SessionService interface {
	// Create creates a new session with the given title.
	Create(ctx context.Context, title string) (Session, error)

	// CreateTaskSession creates a nested session for agent tool execution.
	CreateTaskSession(ctx context.Context, toolCallID, parentSessionID, title string) (Session, error)

	// Get retrieves a session by ID.
	Get(ctx context.Context, id string) (Session, error)

	// List returns all sessions.
	List(ctx context.Context) ([]Session, error)

	// Save persists changes to a session.
	Save(ctx context.Context, session Session) (Session, error)

	// UpdateTitleAndUsage atomically updates the title and usage fields.
	// This is safer than fetching, modifying, and saving the entire session.
	UpdateTitleAndUsage(ctx context.Context, sessionID, title string, promptTokens, completionTokens int64, cost float64) error

	// Delete removes a session.
	Delete(ctx context.Context, id string) error

	// CreateAgentToolSessionID creates a session ID for agent tool sessions.
	CreateAgentToolSessionID(messageID, toolCallID string) string

	// ParseAgentToolSessionID parses an agent tool session ID into its components.
	ParseAgentToolSessionID(sessionID string) (messageID string, toolCallID string, ok bool)

	// IsAgentToolSession checks if a session ID follows the agent tool session format.
	IsAgentToolSession(sessionID string) bool
}

// MessageService defines the interface for message persistence.
// Messages represent individual communications in a conversation,
// including user inputs, assistant responses, and tool interactions.
type MessageService interface {
	// Create creates a new message in a session.
	Create(ctx context.Context, sessionID string, params CreateMessageParams) (Message, error)

	// Update updates an existing message.
	// Used for streaming responses where content is added incrementally.
	Update(ctx context.Context, message Message) error

	// Get retrieves a message by ID.
	Get(ctx context.Context, id string) (Message, error)

	// List returns all messages in a session.
	List(ctx context.Context, sessionID string) ([]Message, error)

	// Delete removes a message.
	Delete(ctx context.Context, id string) error

	// DeleteSessionMessages removes all messages in a session.
	DeleteSessionMessages(ctx context.Context, sessionID string) error
}

// EventBroker defines the interface for pub/sub event distribution.
// Components publish events when state changes, and subscribers
// are notified in real-time.
type EventBroker[T any] interface {
	// Publish publishes an event to all subscribers.
	Publish(event string, data T)

	// Subscribe adds a subscriber function that receives events.
	Subscribe(subscriber func(event string, data T))

	// Clear removes all subscribers.
	Clear()
}

// AgentCall represents the parameters for a single agent invocation.
type AgentCall struct {
	SessionID        string                     // Session identifier
	Prompt           string                     // User input prompt
	ProviderOptions  fantasy.ProviderOptions // Provider-specific options
	Attachments      []Attachment              // File attachments
	MaxOutputTokens  int64                      // Maximum output tokens
	Temperature      *float64                   // Sampling temperature
	TopP             *float64                   // Nucleus sampling parameter
	TopK             *int64                      // Top-k sampling parameter
	FrequencyPenalty *float64                   // Frequency penalty for token selection
	PresencePenalty  *float64                   // Presence penalty for token selection
}

// Session represents a conversation session with metadata and usage tracking.
type Session struct {
	ID               string    // Unique session identifier
	ParentSessionID  string    // Parent session ID (for nested agent tools)
	Title            string    // Human-readable title
	MessageCount     int64     // Number of messages in the session
	PromptTokens     int64     // Total prompt tokens used
	CompletionTokens int64     // Total completion tokens used
	SummaryMessageID string    // ID of the summary message (if exists)
	Cost             float64   // Total cost in USD
	Todos            []Todo    // Tracked tasks
	CreatedAt        int64     // Unix timestamp of creation
	UpdatedAt        int64     // Unix timestamp of last update
}

// Message represents a single message in a conversation.
type Message struct {
	ID               string        // Unique message identifier
	SessionID        string        // Parent session ID
	Role             MessageRole   // Role (user, assistant, tool)
	Parts            []ContentPart // Content parts (text, tool calls, etc.)
	Model            string        // Model name used
	Provider         string        // Provider used
	IsSummaryMessage bool          // Whether this is a summary message
	CreatedAt        int64         // Unix timestamp of creation
	UpdatedAt        int64         // Unix timestamp of last update
}

// MessageRole represents the role of a message sender.
type MessageRole string

const (
	// RoleUser represents a user message.
	RoleUser MessageRole = "user"

	// RoleAssistant represents an AI assistant message.
	RoleAssistant MessageRole = "assistant"

	// RoleTool represents a tool execution result.
	RoleTool MessageRole = "tool"
)

// ContentPart represents a piece of content within a message.
// Messages can contain multiple parts of different types.
type ContentPart interface {
	isContentPart()
}

// Todo represents a tracked task within a session.
type Todo struct {
	Content    string     // Task description
	Status     TodoStatus // Task status
	ActiveForm string     // Active voice description (e.g., "Running task")
}

// TodoStatus represents the status of a todo item.
type TodoStatus string

const (
	// TodoStatusPending indicates that the task is not yet started.
	TodoStatusPending TodoStatus = "pending"

	// TodoStatusInProgress indicates that the task is currently being worked on.
	TodoStatusInProgress TodoStatus = "in_progress"

	// TodoStatusCompleted indicates that the task has been completed.
	TodoStatusCompleted TodoStatus = "completed"
)

// CreateMessageParams represents the parameters for creating a new message.
type CreateMessageParams struct {
	Role             MessageRole   // Message role
	Parts            []ContentPart // Content parts
	Model            string        // Model name
	Provider         string        // Provider name
	IsSummaryMessage bool          // Whether this is a summary message
}

// Attachment represents a file attachment.
type Attachment struct {
	FilePath   string // Path to file
	FileName   string // Original filename
	MimeType   string // MIME type
	Content    []byte // File content
	IsTextFile  bool   // Whether attachment is text content
}

// NewTextContent creates a text content part.
func NewTextContent(text string) ContentPart {
	return TextContent{Text: text}
}

// TextContent represents text content within a message.
type TextContent struct {
	Text string `json:"text"`
}

func (TextContent) isContentPart() {}

// NewReasoningContent creates a reasoning content part.
func NewReasoningContent(text string) ContentPart {
	return ReasoningContent{Text: text}
}

// ReasoningContent represents AI reasoning/thinking content.
type ReasoningContent struct {
	Text      string `json:"text"`
	Signature string `json:"signature,omitempty"`
}

func (ReasoningContent) isContentPart() {}

// NewToolCall creates a tool call content part.
func NewToolCall(id, name, input string, finished bool) ContentPart {
	return ToolCall{
		ID:               id,
		Name:             name,
		Input:            input,
		ProviderExecuted: false,
		Finished:         finished,
	}
}

// ToolCall represents a tool execution request from the AI.
type ToolCall struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Input            string `json:"input"`
	ProviderExecuted bool   `json:"provider_executed"`
	Finished         bool   `json:"finished"`
}

func (ToolCall) isContentPart() {}

// NewToolResult creates a tool result content part.
func NewToolResult(toolCallID, name, content string, isError bool) ContentPart {
	return ToolResult{
		ToolCallID: toolCallID,
		Name:       name,
		Content:    content,
		IsError:    isError,
	}
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	ToolCallID string                 `json:"tool_call_id"`
	Name       string                 `json:"name"`
	Content    string                 `json:"content"`
	IsError    bool                   `json:"is_error"`
	Metadata   map[string]interface{}  `json:"metadata,omitempty"`
}

func (ToolResult) isContentPart() {}

// NewFinish creates a finish content part.
func NewFinish(reason, title, description string) ContentPart {
	return Finish{
		Reason:      reason,
		Title:       title,
		Description: description,
		Time:        0,
	}
}

// Finish represents the completion status of a message.
type Finish struct {
	Reason      string `json:"reason"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Time        int64  `json:"time"`
}

func (Finish) isContentPart() {}

// ContextKeys defines the keys used for propagating state through context.
type ContextKeys struct{}

// contextKey is a custom type to avoid key collisions.
type contextKey string

const (
	// SessionIDContextKey is the key for the session ID in context.
	SessionIDContextKey contextKey = "session_id"

	// MessageIDContextKey is the key for the message ID in context.
	MessageIDContextKey contextKey = "message_id"

	// SupportsImagesContextKey is the key for the image support capability in context.
	SupportsImagesContextKey contextKey = "supports_images"

	// ModelNameContextKey is the key for the model name in context.
	ModelNameContextKey contextKey = "model_name"
)

// GetSessionID retrieves the session ID from context.
func GetSessionID(ctx context.Context) string {
	sessionID := ctx.Value(SessionIDContextKey)
	if sessionID == nil {
		return ""
	}
	s, ok := sessionID.(string)
	if !ok {
		return ""
	}
	return s
}

// GetMessageID retrieves the message ID from context.
func GetMessageID(ctx context.Context) string {
	messageID := ctx.Value(MessageIDContextKey)
	if messageID == nil {
		return ""
	}
	s, ok := messageID.(string)
	if !ok {
		return ""
	}
	return s
}

// GetSupportsImages retrieves whether the model supports images from context.
func GetSupportsImages(ctx context.Context) bool {
	supportsImages := ctx.Value(SupportsImagesContextKey)
	if supportsImages == nil {
		return false
	}
	if supports, ok := supportsImages.(bool); ok {
		return supports
	}
	return false
}

// GetModelName retrieves the model name from context.
func GetModelName(ctx context.Context) string {
	modelName := ctx.Value(ModelNameContextKey)
	if modelName == nil {
		return ""
	}
	s, ok := modelName.(string)
	if !ok {
		return ""
	}
	return s
}

// AgentOptions defines the options for creating an agent.
type AgentOptions struct {
	// LargeModel is the primary model used for task execution.
	LargeModel fantasy.LanguageModel

	// SmallModel is the lightweight model used for auxiliary tasks.
	SmallModel fantasy.LanguageModel

	// SystemPrompt is the system prompt for the agent.
	SystemPrompt string

	// SystemPromptPrefix is prepended to the system prompt.
	SystemPromptPrefix string

	// IsSubAgent indicates whether this is a nested agent (used by agent tools).
	IsSubAgent bool

	// DisableAutoSummarize disables automatic summarization when near the context limit.
	DisableAutoSummarize bool

	// Sessions is the session persistence service.
	Sessions SessionService

	// Messages is the message persistence service.
	Messages MessageService

	// Tools are the available tools for the agent.
	Tools []fantasy.AgentTool

	// EventBroker is the event pub/sub broker.
	EventBroker EventBroker[any]
}

// ErrEmptyPrompt is returned when an empty prompt is provided.
var ErrEmptyPrompt = fmt.Errorf("prompt cannot be empty")

// ErrSessionMissing is returned when a session ID is missing.
var ErrSessionMissing = fmt.Errorf("session ID is required")

// ErrSessionBusy is returned when a session is already processing a request.
var ErrSessionBusy = fmt.Errorf("session is busy")
