package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/revrost/counterspell/internal/models"
)

// Message represents a chat message for agents.
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant", "tool"
	Content string `json:"content"` // Message content
	ToolID  string `json:"tool_id,omitempty"`
}

// AgentBackend is the interface for AI backends.
type AgentBackend interface {
	Chat(ctx context.Context, messages []Message, model string) (string, error)
	GetModel(ctx context.Context) string
}

// OpenAIBackend implements OpenAI API.
type OpenAIBackend struct {
	apiKey string
	model  string
}

func NewOpenAIBackend(apiKey string) *OpenAIBackend {
	return &OpenAIBackend{
		apiKey: apiKey,
		model:  "gpt-4",
	}
}

func (b *OpenAIBackend) Chat(ctx context.Context, messages []Message, model string) (string, error) {
	// Placeholder: Implement actual OpenAI API call
	// For now, return a mock response
	slog.Info("OpenAI Chat called", "model", model, "messages", len(messages))
	return "This is a mock OpenAI response. Integrate actual API client.", nil
}

func (b *OpenAIBackend) GetModel(ctx context.Context) string {
	return b.model
}

// AnthropicBackend implements Anthropic (Claude) API.
type AnthropicBackend struct {
	apiKey string
	model  string
}

func NewAnthropicBackend(apiKey string) *AnthropicBackend {
	return &AnthropicBackend{
		apiKey: apiKey,
		model:  "claude-3-opus-20240229",
	}
}

func (b *AnthropicBackend) Chat(ctx context.Context, messages []Message, model string) (string, error) {
	// Placeholder: Implement actual Anthropic API call
	slog.Info("Anthropic Chat called", "model", model, "messages", len(messages))
	return "This is a mock Anthropic response. Integrate actual API client.", nil
}

func (b *AnthropicBackend) GetModel(ctx context.Context) string {
	return b.model
}

// NativeBackend is a simple rule-based backend.
type NativeBackend struct{}

func NewNativeBackend() *NativeBackend {
	return &NativeBackend{}
}

func (b *NativeBackend) Chat(ctx context.Context, messages []Message, model string) (string, error) {
	slog.Info("Native Chat called", "messages", len(messages))
	return "Native backend: I can help with simple tasks. Connect to an AI backend for full capabilities.", nil
}

func (b *NativeBackend) GetModel(ctx context.Context) string {
	return "native"
}

// AgentService manages agent backends and execution.
type AgentService struct {
	settings *SettingsService
	backend  AgentBackend
}

// NewAgentService creates a new agent service.
func NewAgentService(settings *SettingsService) *AgentService {
	return &AgentService{
		settings: settings,
	}
}

// InitializeBackend initializes the AI backend based on settings.
func (s *AgentService) InitializeBackend(ctx context.Context) error {
	settings, err := s.settings.GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	switch settings.AgentBackend {
	case "openai":
		if settings.OpenAIKey == "" {
			return fmt.Errorf("OpenAI API key not set")
		}
		s.backend = NewOpenAIBackend(settings.OpenAIKey)
	case "anthropic":
		if settings.AnthropicKey == "" {
			return fmt.Errorf("anthropic API key not set")
		}
		s.backend = NewAnthropicBackend(settings.AnthropicKey)
	case "openrouter":
		// OpenRouter uses OpenAI-compatible API
		if settings.OpenRouterKey == "" {
			return fmt.Errorf("OpenRouter API key not set")
		}
		s.backend = NewOpenAIBackend(settings.OpenRouterKey)
	case "zai":
		// Zai (placeholder)
		if settings.ZaiKey == "" {
			return fmt.Errorf("zai API key not set")
		}
		s.backend = NewNativeBackend() // Fall back to native for now
	case "native":
		s.backend = NewNativeBackend()
	default:
		return fmt.Errorf("unknown agent_backend: %s", settings.AgentBackend)
	}

	slog.Info("Agent backend initialized", "backend", settings.AgentBackend)
	return nil
}

// ExecuteTask executes a task using the agent backend.
func (s *AgentService) ExecuteTask(ctx context.Context, task *models.Task) (string, error) {
	if s.backend == nil {
		if err := s.InitializeBackend(ctx); err != nil {
			return "", err
		}
	}

	// Build message history
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful coding assistant. Analyze the user's request and provide solutions.",
		},
		{
			Role:    "user",
			Content: task.Intent,
		},
	}

	// Call the backend
	response, err := s.backend.Chat(ctx, messages, "")
	if err != nil {
		return "", fmt.Errorf("agent chat failed: %w", err)
	}

	return response, nil
}

// ExecuteWithTool calls the agent with tool calls.
func (s *AgentService) ExecuteWithTool(ctx context.Context, taskID string, messages []Message, tools []string) (string, error) {
	if s.backend == nil {
		if err := s.InitializeBackend(ctx); err != nil {
			return "", err
		}
	}

	// Add tools to system message
	toolsJSON, _ := json.Marshal(tools)
	systemContent := fmt.Sprintf("You are a helpful coding assistant with access to these tools: %s. Use them to accomplish tasks.", string(toolsJSON))

	messages = append([]Message{
		{
			Role:    "system",
			Content: systemContent,
		},
	}, messages...)

	// Call the backend
	response, err := s.backend.Chat(ctx, messages, "")
	if err != nil {
		return "", fmt.Errorf("agent chat with tools failed: %w", err)
	}

	return response, nil
}

// StreamTask executes a task with streaming support.
func (s *AgentService) StreamTask(ctx context.Context, task *models.Task, callback func(chunk string)) error {
	if s.backend == nil {
		if err := s.InitializeBackend(ctx); err != nil {
			return err
		}
	}

	// Build messages
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful coding assistant.",
		},
		{
			Role:    "user",
			Content: task.Intent,
		},
	}

	// For streaming, we'll simulate chunks
	// In production, implement actual streaming based on backend
	response, err := s.backend.Chat(ctx, messages, "")
	if err != nil {
		return err
	}

	// Simulate streaming by sending chunks
	chunks := splitIntoChunks(response, 20)
	for _, chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond): // Simulate delay
			callback(chunk)
		}
	}

	return nil
}

// splitIntoChunks splits text into chunks for streaming.
func splitIntoChunks(text string, chunkSize int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	for i := 0; i < len(text); i += chunkSize {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}
