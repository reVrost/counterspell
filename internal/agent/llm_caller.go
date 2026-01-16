package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/revrost/code/counterspell/internal/agent/tools"
	"github.com/revrost/code/counterspell/internal/llm"
)

const (
	// apiURL   = "https://api.anthropic.com/v1/messages"
	// model    = "claude-opus-4-5"
	// apiVer   = "2023-06-01"
	maxToken = 8192
)

// APIRequest is what we send to Anthropic's API.
type APIRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system"`   // System prompt (context for the assistant)
	Messages  []Message       `json:"messages"` // Conversation history
	Tools     []tools.ToolDef `json:"tools"`    // Tools available to the assistant
}

// APIResponse is what we get back from Anthropic's API.
type APIResponse struct {
	Content []ContentBlock `json:"content"` // Assistant's response
}

// LLMCaller is an interface for calling LLM APIs.
// Implementations handle the specific protocol (Anthropic vs OpenAI).
type LLMCaller interface {
	Call(messages []Message, allTools map[string]tools.Tool, systemPrompt string) (*APIResponse, error)
}

// NewLLMCaller creates an LLMCaller based on the provider type.
func NewLLMCaller(provider llm.Provider) LLMCaller {
	switch provider.Type() {
	case "openai":
		return &OpenAICaller{provider: provider}
	default:
		return &AnthropicCaller{provider: provider}
	}
}

// AnthropicCaller implements LLMCaller for Anthropic-compatible APIs.
type AnthropicCaller struct {
	provider llm.Provider
}

func (c *AnthropicCaller) Call(messages []Message, allTools map[string]tools.Tool, systemPrompt string) (*APIResponse, error) {
	req := APIRequest{
		Model:     c.provider.Model(),
		MaxTokens: maxToken,
		System:    systemPrompt,
		Messages:  messages,
		Tools:     tools.MakeSchema(allTools),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	prettyBody, _ := json.MarshalIndent(req, "", "  ")
	slog.Info("[LLM REQUEST] Sending to API",
		"url", c.provider.APIURL(),
		"model", c.provider.Model(),
		"message_count", len(messages),
		"tool_count", len(allTools),
	)
	slog.Debug("[LLM REQUEST] Full payload", "body", string(prettyBody))

	httpReq, err := http.NewRequest("POST", c.provider.APIURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	providerType := detectProviderType(c.provider.APIURL())
	switch providerType {
	case "anthropic":
		httpReq.Header.Set("x-api-key", c.provider.APIKey())
		httpReq.Header.Set("anthropic-version", c.provider.APIVersion())
	case "openrouter":
		httpReq.Header.Set("Authorization", "Bearer "+c.provider.APIKey())
		httpReq.Header.Set("HTTP-Referer", "https://counterspell.dev")
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &apiResp, nil
}

// OpenAICaller implements LLMCaller for OpenAI-compatible APIs.
type OpenAICaller struct {
	provider llm.Provider
}

// OpenAI-specific request/response types

type OpenAIRequest struct {
	Model      string          `json:"model"`
	Messages   []OpenAIMessage `json:"messages"`
	Tools      []OpenAIToolDef `json:"tools,omitempty"`
	ToolChoice string          `json:"tool_choice,omitempty"`
}

type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    any              `json:"content"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type OpenAIContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type OpenAIToolDef struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

type FunctionDef struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  tools.InputSchema `json:"parameters"`
}

type OpenAIToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message OpenAIMessage `json:"message"`
	} `json:"choices"`
}

func (c *OpenAICaller) Call(messages []Message, allTools map[string]tools.Tool, systemPrompt string) (*APIResponse, error) {
	openAIMessages := []OpenAIMessage{}

	openAIMessages = append(openAIMessages, OpenAIMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, msg := range messages {
		if msg.Role == "user" {
			isToolResult := false
			for _, block := range msg.Content {
				if block.Type == "tool_result" {
					isToolResult = true
					openAIMessages = append(openAIMessages, OpenAIMessage{
						Role:       "tool",
						ToolCallID: block.ToolUseID,
						Content:    block.Content,
					})
				}
			}

			if !isToolResult {
				var contentBuilder strings.Builder
				for _, block := range msg.Content {
					if block.Type == "text" {
						contentBuilder.WriteString(block.Text)
					}
				}
				openAIMessages = append(openAIMessages, OpenAIMessage{
					Role:    "user",
					Content: contentBuilder.String(),
				})
			}
		}

		if msg.Role == "assistant" {
			oaMsg := OpenAIMessage{
				Role: "assistant",
			}

			var contentBuilder strings.Builder
			for _, block := range msg.Content {
				if block.Type == "text" {
					contentBuilder.WriteString(block.Text)
				}
				if block.Type == "tool_use" {
					argsJSON, _ := json.Marshal(block.Input)
					oaMsg.ToolCalls = append(oaMsg.ToolCalls, OpenAIToolCall{
						ID:   block.ID,
						Type: "function",
						Function: FunctionCall{
							Name:      block.Name,
							Arguments: string(argsJSON),
						},
					})
				}
			}
			oaMsg.Content = contentBuilder.String()
			openAIMessages = append(openAIMessages, oaMsg)
		}
	}

	openAITools := []OpenAIToolDef{}
	for name, tool := range allTools {
		openAITools = append(openAITools, OpenAIToolDef{
			Type: "function",
			Function: FunctionDef{
				Name:        name,
				Description: tool.Description,
				Parameters:  tools.MakeSchema(map[string]tools.Tool{name: tool})[0].InputSchema,
			},
		})
	}

	req := OpenAIRequest{
		Model:    c.provider.Model(),
		Messages: openAIMessages,
		Tools:    openAITools,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.provider.APIURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.provider.APIKey())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	var oaResp OpenAIResponse
	if err := json.Unmarshal(respBody, &oaResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	apiResp := APIResponse{
		Content: []ContentBlock{},
	}

	if len(oaResp.Choices) > 0 {
		choice := oaResp.Choices[0]

		if contentStr, ok := choice.Message.Content.(string); ok && contentStr != "" {
			apiResp.Content = append(apiResp.Content, ContentBlock{
				Type: "text",
				Text: contentStr,
			})
		}

		for _, toolCall := range choice.Message.ToolCalls {
			var input map[string]any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
				continue
			}

			apiResp.Content = append(apiResp.Content, ContentBlock{
				Type:  "tool_use",
				ID:    toolCall.ID,
				Name:  toolCall.Function.Name,
				Input: input,
			})
		}
	}

	return &apiResp, nil
}

// detectProviderType determines the provider type from the API URL.
func detectProviderType(apiURL string) string {
	if strings.Contains(apiURL, "anthropic") {
		return "anthropic"
	}
	if strings.Contains(apiURL, "openrouter") {
		return "openrouter"
	}
	if strings.Contains(apiURL, "z.ai") {
		return "zai"
	}
	return ""
}
