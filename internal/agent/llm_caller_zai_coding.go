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

// ZAICodingCaller implements LLMCaller for Z.AI Coding PaaS endpoint.
type ZAICodingCaller struct {
	provider llm.Provider
}

// Z.AI coding request uses OpenAI-compatible chat/completions with a few extensions.
type ZAIChatRequest struct {
	Model      string          `json:"model"`
	Messages   []OpenAIMessage `json:"messages"`
	Tools      []OpenAIToolDef `json:"tools,omitempty"`
	ToolChoice string          `json:"tool_choice,omitempty"`
	Stream     bool            `json:"stream,omitempty"`
	ToolStream bool            `json:"tool_stream,omitempty"`
}

func (c *ZAICodingCaller) Stream(ctx context.Context, messages []Message, allTools map[string]tools.Tool, systemPrompt string) (*LLMStream, error) {
	supportsTools := true
	req := ZAIChatRequest{
		Model:      c.provider.Model(),
		Messages:   toOpenAIMessages(messages, systemPrompt, supportsTools),
		Tools:      toOpenAITools(allTools, supportsTools),
		ToolChoice: "auto",
		Stream:     true,
		ToolStream: true,
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
		toolActive := map[int]OpenAIToolCall{}

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
						Content          string `json:"content"`
						ReasoningContent string `json:"reasoning_content"`
						Reasoning        string `json:"reasoning"`
						ToolCalls        []struct {
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
				reasoning := choice.Delta.Reasoning
				if reasoning == "" {
					reasoning = choice.Delta.ReasoningContent
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

				if choice.Delta.Content != "" {
					if thinkingActive {
						emit(LLMEvent{Type: LLMContentEnd, BlockType: "thinking"})
						thinkingActive = false
					}
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
					if thinkingActive {
						emit(LLMEvent{Type: LLMContentEnd, BlockType: "thinking"})
						thinkingActive = false
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
