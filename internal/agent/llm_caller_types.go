package agent

import (
	"context"

	"github.com/revrost/counterspell/internal/agent/tools"
	"github.com/revrost/counterspell/internal/llm"
)

const (
	maxToken = 8192
)

// APIRequest is what we send to Anthropic-compatible APIs.
type APIRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system"`
	Messages  []Message       `json:"messages"`
	Tools     []tools.ToolDef `json:"tools"`
	Stream    bool            `json:"stream,omitempty"`
}

// LLMEventType identifies the type of streaming event from the LLM.
type LLMEventType string

const (
	LLMContentStart LLMEventType = "content_start"
	LLMContentDelta LLMEventType = "content_delta"
	LLMContentEnd   LLMEventType = "content_end"
	LLMMessageEnd   LLMEventType = "message_end"
)

// LLMEvent represents a single streaming event from the LLM.
type LLMEvent struct {
	Type      LLMEventType
	BlockType string
	Delta     string
	Block     *ContentBlock
}

// LLMStream represents an asynchronous stream of LLM events.
type LLMStream struct {
	Events <-chan LLMEvent
	Done   <-chan error
}

// LLMCaller is an interface for calling LLM APIs.
type LLMCaller interface {
	Stream(ctx context.Context, messages []Message, allTools map[string]tools.Tool, systemPrompt string) (*LLMStream, error)
}

// NewLLMCaller creates an LLMCaller based on the provider type.
func NewLLMCaller(provider llm.Provider) LLMCaller {
	switch provider.Type() {
	case "openrouter":
		return &OpenRouterCaller{provider: provider}
	case "zai-coding":
		return &ZAICodingCaller{provider: provider}
	case "openai":
		return &OpenAICaller{provider: provider}
	default:
		return &AnthropicCaller{provider: provider}
	}
}
