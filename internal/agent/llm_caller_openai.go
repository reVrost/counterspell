package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/revrost/counterspell/internal/agent/tools"
	"github.com/revrost/counterspell/internal/llm"
)

// OpenAICaller implements LLMCaller for OpenAI-compatible APIs.
type OpenAICaller struct {
	provider llm.Provider
}

// OpenAI-specific request/response types.
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
	supportsTools := true
	openAIMessages := toOpenAIMessages(messages, systemPrompt, supportsTools)
	openAITools := toOpenAITools(allTools, supportsTools)

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
						toolID := tc.ID
						if toolID == "" {
							toolID = fmt.Sprintf("call_%d", tc.Index)
						}
						toolActive[tc.Index] = OpenAIToolCall{ID: toolID, Function: FunctionCall{Name: tc.Function.Name}}
						emit(LLMEvent{Type: LLMContentStart, BlockType: "tool_use", Block: &ContentBlock{Type: "tool_use", Name: tc.Function.Name, ID: toolID}})
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

func toOpenAIMessages(messages []Message, systemPrompt string, supportsTools bool) []OpenAIMessage {
	openAIMessages := []OpenAIMessage{{Role: "system", Content: systemPrompt}}

	for _, msg := range messages {
		isToolResult := false
		for _, block := range msg.Content {
			if block.Type == "tool_result" {
				isToolResult = true
				if supportsTools {
					openAIMessages = append(openAIMessages, OpenAIMessage{
						Role:       "tool",
						ToolCallID: block.ToolUseID,
						Content:    block.Content,
					})
				}
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
			openAIMessages = append(openAIMessages, OpenAIMessage{Role: "user", Content: contentBuilder.String()})
		}

		if msg.Role == "assistant" {
			oaMsg := OpenAIMessage{Role: "assistant"}
			var contentBuilder strings.Builder
			for _, block := range msg.Content {
				if block.Type == "text" {
					contentBuilder.WriteString(block.Text)
				}
				if block.Type == "tool_use" && supportsTools {
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

	return openAIMessages
}

func toOpenAITools(allTools map[string]tools.Tool, supportsTools bool) []OpenAIToolDef {
	if !supportsTools {
		return nil
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
	return openAITools
}
