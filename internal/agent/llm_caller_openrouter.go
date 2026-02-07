package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/revrost/counterspell/internal/agent/tools"
	"github.com/revrost/counterspell/internal/llm"
	openrouter "github.com/revrost/go-openrouter"
)

// OpenRouterCaller implements LLMCaller via github.com/revrost/go-openrouter.
type OpenRouterCaller struct {
	provider llm.Provider
}

func (c *OpenRouterCaller) Stream(ctx context.Context, messages []Message, allTools map[string]tools.Tool, systemPrompt string) (*LLMStream, error) {
	cfg := openrouter.DefaultConfig(c.provider.APIKey())
	baseURL := strings.TrimSpace(c.provider.APIURL())
	baseURL = strings.TrimSuffix(baseURL, "/chat/completions")
	baseURL = strings.TrimSuffix(baseURL, "/messages")
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	client := openrouter.NewClientWithConfig(*cfg)

	req := openrouter.ChatCompletionRequest{
		Model:    c.provider.Model(),
		Messages: toOpenRouterMessages(messages, systemPrompt),
		Tools:    toOpenRouterTools(allTools),
		Stream:   true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openrouter stream start: %w", err)
	}

	events := make(chan LLMEvent, 32)
	done := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(done)
		defer stream.Close()

		emit := func(ev LLMEvent) bool {
			select {
			case <-ctx.Done():
				return false
			case events <- ev:
				return true
			}
		}

		textActive := false
		thinkingActive := false
		toolActive := map[int]openrouter.ToolCall{}
		nextToolIndex := 0

		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				done <- err
				return
			}

			for _, choice := range resp.Choices {
				delta := choice.Delta

				if delta.Content != "" {
					if thinkingActive {
						emit(LLMEvent{Type: LLMContentEnd, BlockType: "thinking"})
						thinkingActive = false
					}
					if !textActive {
						emit(LLMEvent{Type: LLMContentStart, BlockType: "text", Block: &ContentBlock{Type: "text"}})
						textActive = true
					}
					emit(LLMEvent{Type: LLMContentDelta, BlockType: "text", Delta: delta.Content})
				}

				reasoning := ""
				if delta.Reasoning != nil {
					reasoning = *delta.Reasoning
				}
				if reasoning == "" {
					reasoning = delta.ReasoningContent
				}
				if reasoning != "" {
					if textActive {
						emit(LLMEvent{Type: LLMContentEnd, BlockType: "text"})
						textActive = false
					}
					if !thinkingActive {
						emit(LLMEvent{Type: LLMContentStart, BlockType: "thinking", Block: &ContentBlock{Type: "thinking"}})
						thinkingActive = true
					}
					emit(LLMEvent{Type: LLMContentDelta, BlockType: "thinking", Delta: reasoning})
				}

				for _, tc := range delta.ToolCalls {
					idx := nextToolIndex
					if tc.Index != nil {
						idx = *tc.Index
					}
					if idx >= nextToolIndex {
						nextToolIndex = idx + 1
					}

					call, ok := toolActive[idx]
					if !ok {
						call = tc
						if call.ID == "" {
							call.ID = fmt.Sprintf("call_%d", idx)
						}
						toolActive[idx] = call
						emit(LLMEvent{Type: LLMContentStart, BlockType: "tool_use", Block: &ContentBlock{Type: "tool_use", Name: call.Function.Name, ID: call.ID}})
					}
					if tc.Function.Arguments != "" {
						emit(LLMEvent{Type: LLMContentDelta, BlockType: "tool_use", Delta: tc.Function.Arguments})
					}
				}

				if choice.FinishReason != "" {
					if textActive {
						emit(LLMEvent{Type: LLMContentEnd, BlockType: "text"})
						textActive = false
					}
					if thinkingActive {
						emit(LLMEvent{Type: LLMContentEnd, BlockType: "thinking"})
						thinkingActive = false
					}
					if len(toolActive) > 0 {
						for range toolActive {
							emit(LLMEvent{Type: LLMContentEnd, BlockType: "tool_use"})
						}
						toolActive = map[int]openrouter.ToolCall{}
					}
				}
			}
		}

		if textActive {
			emit(LLMEvent{Type: LLMContentEnd, BlockType: "text"})
		}
		if thinkingActive {
			emit(LLMEvent{Type: LLMContentEnd, BlockType: "thinking"})
		}
		if len(toolActive) > 0 {
			for range toolActive {
				emit(LLMEvent{Type: LLMContentEnd, BlockType: "tool_use"})
			}
		}
		emit(LLMEvent{Type: LLMMessageEnd})
		done <- nil
	}()

	return &LLMStream{Events: events, Done: done}, nil
}

func toOpenRouterMessages(messages []Message, systemPrompt string) []openrouter.ChatCompletionMessage {
	orMsgs := []openrouter.ChatCompletionMessage{{
		Role:    "system",
		Content: openrouter.Content{Text: systemPrompt},
	}}

	for _, msg := range messages {
		isToolResult := false
		for _, block := range msg.Content {
			if block.Type != "tool_result" {
				continue
			}
			isToolResult = true
			orMsgs = append(orMsgs, openrouter.ChatCompletionMessage{
				Role:       "tool",
				ToolCallID: block.ToolUseID,
				Content:    openrouter.Content{Text: block.Content},
			})
		}
		if isToolResult {
			continue
		}

		switch msg.Role {
		case "user":
			var contentBuilder strings.Builder
			for _, block := range msg.Content {
				if block.Type == "text" {
					contentBuilder.WriteString(block.Text)
				}
			}
			orMsgs = append(orMsgs, openrouter.ChatCompletionMessage{
				Role:    "user",
				Content: openrouter.Content{Text: contentBuilder.String()},
			})
		case "assistant":
			var contentBuilder strings.Builder
			assistant := openrouter.ChatCompletionMessage{Role: "assistant"}
			for _, block := range msg.Content {
				if block.Type == "text" {
					contentBuilder.WriteString(block.Text)
				}
				if block.Type == "tool_use" {
					argsJSON, _ := json.Marshal(block.Input)
					assistant.ToolCalls = append(assistant.ToolCalls, openrouter.ToolCall{
						ID:   block.ID,
						Type: openrouter.ToolTypeFunction,
						Function: openrouter.FunctionCall{
							Name:      block.Name,
							Arguments: string(argsJSON),
						},
					})
				}
			}
			assistant.Content = openrouter.Content{Text: contentBuilder.String()}
			orMsgs = append(orMsgs, assistant)
		}
	}

	return orMsgs
}

func toOpenRouterTools(allTools map[string]tools.Tool) []openrouter.Tool {
	orTools := make([]openrouter.Tool, 0, len(allTools))
	for name, tool := range allTools {
		orTools = append(orTools, openrouter.Tool{
			Type: openrouter.ToolTypeFunction,
			Function: &openrouter.FunctionDefinition{
				Name:        name,
				Description: tool.Description,
				Parameters:  tools.MakeSchema(map[string]tools.Tool{name: tool})[0].InputSchema,
			},
		})
	}
	return orTools
}
