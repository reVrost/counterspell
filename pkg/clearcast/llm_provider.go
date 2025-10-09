package clearcast

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/revrost/go-openrouter"
)

// LLMProvider defines the minimal interface for any chat LLM backend.
type LLMProvider interface {
	ChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error)
	ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (<-chan ChatCompletionChunk, error)
}

// ChatCompletionRequest is a provider-neutral request type.
type ChatCompletionRequest struct {
	Model          string
	Messages       []ChatMessage
	Options        map[string]any
	ResponseFormat *ResponseFormat
}

type ResponseFormatType string

const (
	ResponseFormatTypeJSON       = "json_object"
	ResponseFormatTypeJSONSchema = "json_schema"
	ResponseFormatTypeText       = "text"
)

type ResponseFormat struct {
	Type       ResponseFormatType
	JSONSchema *JSONSchema
}

type JSONSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      json.Marshaler `json:"schema"`
	Strict      bool           `json:"strict"`
}

// ChatCompletionResponse represents a full completion result.
type ChatCompletionResponse struct {
	Content string
	Raw     any // keep raw provider response if needed
	Usage   Usage
}

// ChatCompletionChunk is a streamed chunk of a completion.
type ChatCompletionChunk struct {
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

func (p *OpenRouterProvider) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	slog.Debug("calling OpenRouter provider with request", "request", req)
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

	if req.ResponseFormat != nil {
		orReq.ResponseFormat = &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatType(req.ResponseFormat.Type),
		}
		if req.ResponseFormat.JSONSchema != nil {
			orReq.ResponseFormat.JSONSchema = &openrouter.ChatCompletionResponseFormatJSONSchema{
				Name:        req.ResponseFormat.JSONSchema.Name,
				Description: req.ResponseFormat.JSONSchema.Description,
				Schema:      req.ResponseFormat.JSONSchema.Schema,
				Strict:      req.ResponseFormat.JSONSchema.Strict,
			}
		}
	}

	resp, err := p.client.CreateChatCompletion(ctx, orReq)
	if err != nil {
		return ChatCompletionResponse{}, err
	}

	return ChatCompletionResponse{
		Content: resp.Choices[0].Message.Content.Text,
		Raw:     resp,
	}, nil
}

func (p *OpenRouterProvider) ChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (<-chan ChatCompletionChunk, error) {
	// Map openrouter's streaming API into your chunk channel here.
	slog.Info("stream calling OpenRouter provider with request", "request", req)

	messages := make([]openrouter.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openrouter.ChatCompletionMessage{
			Role:    msg.Role,
			Content: openrouter.Content{Text: msg.Content},
		}
	}
	stream, err := p.client.CreateChatCompletionStream(
		context.Background(), openrouter.ChatCompletionRequest{
			Model:    req.Model,
			Messages: messages,
			Stream:   true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating chat completion stream: %w", err)
	}
	defer stream.Close()

	responseChan := make(chan ChatCompletionChunk)

	go func() {
		for {
			response, err := stream.Recv()
			slog.Info("Received response", "response", response)
			if errors.Is(err, io.EOF) {
				slog.Warn("Stream closed, exiting")
				break
			}
			responseChan <- ChatCompletionChunk{
				Delta: response.Choices[0].Delta.Content,
				Done:  response.Choices[0].FinishReason == "stop",
				Raw:   response,
			}
		}
	}()

	return responseChan, nil
}

func DecodeMessage[T any](msg ChatCompletionResponse) (T, error) {
	var data T
	err := json.Unmarshal([]byte(msg.Content), &data)
	return data, err
}
