package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/revrost/counterspell/internal/agent/tools"
	"github.com/revrost/counterspell/internal/llm"
)

const (
	maxToken = 8192
)

// APIRequest is what we send to Anthropic's API.
type APIRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system"`
	Messages  []Message       `json:"messages"`
	Tools     []tools.ToolDef `json:"tools"`
	Stream    bool            `json:"stream,omitempty"`
}

// APIResponse is what we get back from Anthropic's API.
type APIResponse struct {
	Content []ContentBlock `json:"content"`
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

func (c *AnthropicCaller) Stream(ctx context.Context, messages []Message, allTools map[string]tools.Tool, systemPrompt string) (*LLMStream, error) {
	req := APIRequest{
		Model:     c.provider.Model(),
		MaxTokens: maxToken,
		System:    systemPrompt,
		Messages:  messages,
		Tools:     tools.MakeSchema(allTools),
		Stream:    true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	prettyBody, _ := json.MarshalIndent(req, "", "  ")
	slog.Info("[LLM STREAM] Sending to API",
		"url", c.provider.APIURL(),
		"model", c.provider.Model(),
		"message_count", len(messages),
		"tool_count", len(allTools),
	)
	slog.Debug("[LLM STREAM] Full payload", "body", string(prettyBody))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.provider.APIURL(), bytes.NewReader(body))
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

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	events := make(chan LLMEvent, 32)
	done := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(done)
		defer resp.Body.Close()

		blockTypes := map[int]string{}
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		var eventName string
		var data strings.Builder

		emit := func(ev LLMEvent) bool {
			select {
			case <-ctx.Done():
				return false
			case events <- ev:
				return true
			}
		}

		flush := func() bool {
			payload := strings.TrimSpace(data.String())
			name := eventName
			eventName = ""
			data.Reset()
			if payload == "" {
				return true
			}

			switch name {
			case "content_block_start":
				var evt struct {
					Index        int `json:"index"`
					ContentBlock struct {
						Type     string         `json:"type"`
						Text     string         `json:"text,omitempty"`
						Thinking string         `json:"thinking,omitempty"`
						Name     string         `json:"name,omitempty"`
						ID       string         `json:"id,omitempty"`
						Input    map[string]any `json:"input,omitempty"`
					} `json:"content_block"`
				}
				if err := json.Unmarshal([]byte(payload), &evt); err != nil {
					return true
				}
				blockTypes[evt.Index] = evt.ContentBlock.Type
				block := &ContentBlock{Type: evt.ContentBlock.Type}
				switch evt.ContentBlock.Type {
				case "text":
					block.Text = evt.ContentBlock.Text
				case "thinking":
					block.Text = evt.ContentBlock.Thinking
				case "tool_use":
					block.Name = evt.ContentBlock.Name
					block.ID = evt.ContentBlock.ID
					block.Input = evt.ContentBlock.Input
				}
				return emit(LLMEvent{Type: LLMContentStart, BlockType: evt.ContentBlock.Type, Block: block})
			case "content_block_delta":
				var evt struct {
					Index int `json:"index"`
					Delta struct {
						Type        string `json:"type"`
						Text        string `json:"text,omitempty"`
						Thinking    string `json:"thinking,omitempty"`
						PartialJSON string `json:"partial_json,omitempty"`
					} `json:"delta"`
				}
				if err := json.Unmarshal([]byte(payload), &evt); err != nil {
					return true
				}
				if evt.Delta.Text != "" {
					return emit(LLMEvent{Type: LLMContentDelta, BlockType: "text", Delta: evt.Delta.Text})
				}
				if evt.Delta.Thinking != "" {
					return emit(LLMEvent{Type: LLMContentDelta, BlockType: "thinking", Delta: evt.Delta.Thinking})
				}
				if evt.Delta.PartialJSON != "" {
					return emit(LLMEvent{Type: LLMContentDelta, BlockType: "tool_use", Delta: evt.Delta.PartialJSON})
				}
			case "content_block_stop":
				var evt struct {
					Index int `json:"index"`
				}
				if err := json.Unmarshal([]byte(payload), &evt); err != nil {
					return true
				}
				blockType := blockTypes[evt.Index]
				return emit(LLMEvent{Type: LLMContentEnd, BlockType: blockType})
			case "message_stop":
				return emit(LLMEvent{Type: LLMMessageEnd})
			case "error":
				var evt struct {
					Error struct {
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.Unmarshal([]byte(payload), &evt); err == nil {
					done <- fmt.Errorf("llm error: %s", evt.Error.Message)
					return false
				}
			}
			return true
		}

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				if !flush() {
					return
				}
				continue
			}
			if strings.HasPrefix(line, "event:") {
				eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
				continue
			}
			if strings.HasPrefix(line, "data:") {
				if data.Len() > 0 {
					data.WriteByte('\n')
				}
				data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			}
		}

		if err := scanner.Err(); err != nil {
			done <- err
			return
		}
		done <- nil
	}()

	return &LLMStream{Events: events, Done: done}, nil
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
	Stream     bool            `json:"stream,omitempty"`
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

func (c *OpenAICaller) Stream(ctx context.Context, messages []Message, allTools map[string]tools.Tool, systemPrompt string) (*LLMStream, error) {
	providerType := detectProviderType(c.provider.APIURL())
	supportsTools := providerType != "zai"

	openAIMessages := []OpenAIMessage{}

	openAIMessages = append(openAIMessages, OpenAIMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, msg := range messages {
		isToolResult := false
		for _, block := range msg.Content {
			if block.Type == "tool_result" {
				if !supportsTools {
					isToolResult = true
					continue
				}
				isToolResult = true
				openAIMessages = append(openAIMessages, OpenAIMessage{
					Role:       "tool",
					ToolCallID: block.ToolUseID,
					Content:    block.Content,
				})
			}
		}

		if isToolResult {
			continue
		}

		if msg.Role == "user" {
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

		if msg.Role == "assistant" {
			oaMsg := OpenAIMessage{Role: "assistant"}
			var contentBuilder strings.Builder
			for _, block := range msg.Content {
				if block.Type == "text" {
					contentBuilder.WriteString(block.Text)
				}
				if block.Type == "tool_use" {
					if !supportsTools {
						continue
					}
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
			if !supportsTools && contentBuilder.Len() == 0 {
				continue
			}
			oaMsg.Content = contentBuilder.String()
			openAIMessages = append(openAIMessages, oaMsg)
		}
	}

	openAITools := []OpenAIToolDef{}
	if supportsTools {
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
	}

	req := OpenAIRequest{
		Model:    c.provider.Model(),
		Messages: openAIMessages,
		Tools:    openAITools,
		Stream:   true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.provider.APIURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.provider.APIKey())

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	events := make(chan LLMEvent, 32)
	done := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(done)
		defer resp.Body.Close()

		textActive := false
		toolActive := map[int]OpenAIToolCall{}

		emit := func(ev LLMEvent) bool {
			select {
			case <-ctx.Done():
				return false
			case events <- ev:
				return true
			}
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				break
			}

			var payload struct {
				Choices []struct {
					Delta struct {
						Content   string `json:"content"`
						ToolCalls []struct {
							Index    int    `json:"index"`
							ID       string `json:"id"`
							Type     string `json:"type"`
							Function struct {
								Name      string `json:"name"`
								Arguments string `json:"arguments"`
							} `json:"function"`
						} `json:"tool_calls"`
					} `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				continue
			}

			for _, choice := range payload.Choices {
				if choice.Delta.Content != "" {
					if !textActive {
						emit(LLMEvent{Type: LLMContentStart, BlockType: "text", Block: &ContentBlock{Type: "text"}})
						textActive = true
					}
					emit(LLMEvent{Type: LLMContentDelta, BlockType: "text", Delta: choice.Delta.Content})
				}

				for _, tc := range choice.Delta.ToolCalls {
					if _, ok := toolActive[tc.Index]; !ok {
						toolActive[tc.Index] = OpenAIToolCall{ID: tc.ID, Function: FunctionCall{Name: tc.Function.Name}}
						emit(LLMEvent{Type: LLMContentStart, BlockType: "tool_use", Block: &ContentBlock{Type: "tool_use", Name: tc.Function.Name, ID: tc.ID}})
					}
					if tc.Function.Arguments != "" {
						emit(LLMEvent{Type: LLMContentDelta, BlockType: "tool_use", Delta: tc.Function.Arguments})
					}
				}

				if choice.FinishReason != nil {
					if textActive {
						emit(LLMEvent{Type: LLMContentEnd, BlockType: "text"})
						textActive = false
					}
					if len(toolActive) > 0 {
						for range toolActive {
							emit(LLMEvent{Type: LLMContentEnd, BlockType: "tool_use"})
						}
						toolActive = map[int]OpenAIToolCall{}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			done <- err
			return
		}
		emit(LLMEvent{Type: LLMMessageEnd})
		done <- nil
	}()

	return &LLMStream{Events: events, Done: done}, nil
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
