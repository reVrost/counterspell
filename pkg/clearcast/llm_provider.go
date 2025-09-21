package clearcast

import (
	"context"
	"fmt"

	"github.com/revrost/go-openrouter"
)

// LLMProvider defines the minimal interface for any chat LLM backend.
type LLMProvider interface {
	ChatCompletion(ctx context.Context, req ChatCompletionRequest) (RunResponse, error)
	ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (<-chan RunChunk, error)
}

// ChatCompletionRequest is a provider-neutral request type.
type ChatCompletionRequest struct {
	Model    string
	Messages []ChatMessage
	Options  map[string]any // extensible for provider-specific options
}

// RunResponse represents a full completion result.
type RunResponse struct {
	Content string
	Raw     any // keep raw provider response if needed
	Usage   Usage
}

// RunChunk is a streamed chunk of a completion.
type RunChunk struct {
	Delta string // partial text delta
	Done  bool
	Raw   any
	Usage Usage
}

// Usage Represents the total token usage per request.
type Usage struct {
	PromptTokens           int                    `json:"prompt_tokens"`
	CompletionTokens       int                    `json:"completion_tokens"`
	CompletionTokenDetails CompletionTokenDetails `json:"completion_token_details"`
	TotalTokens            int                    `json:"total_tokens"`
	Cost                   float64                `json:"cost"`
	CostDetails            CostDetails            `json:"cost_details"`
	PromptTokenDetails     PromptTokenDetails     `json:"prompt_token_details"`
}

type CostDetails struct {
	UpstreamInferenceCost float64 `json:"upstream_inference_cost"`
}

type CompletionTokenDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type PromptTokenDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type OpenRouterProvider struct {
	client *openrouter.Client
}

func NewOpenRouterProvider(client *openrouter.Client) *OpenRouterProvider {
	return &OpenRouterProvider{client: client}
}

func (p *OpenRouterProvider) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (RunResponse, error) {
	messages := make([]openrouter.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openrouter.ChatCompletionMessage{
			Role:    msg.Role,
			Content: openrouter.Content{Text: msg.Content},
		}
	}

	orReq := openrouter.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
	}

	resp, err := p.client.CreateChatCompletion(ctx, orReq)
	if err != nil {
		return RunResponse{}, err
	}

	return RunResponse{
		Content: resp.Choices[0].Message.Content.Text,
		Raw:     resp,
	}, nil
}

func (p *OpenRouterProvider) ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (<-chan RunChunk, error) {
	// Map openrouter's streaming API into your chunk channel here.
	return nil, fmt.Errorf("streaming not implemented yet")
}
