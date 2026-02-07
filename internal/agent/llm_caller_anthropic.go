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

	slog.Info("[LLM STREAM] Sending to API",
		"provider", c.provider.Type(),
		"url", c.provider.APIURL(),
		"model", c.provider.Model(),
		"message_count", len(messages),
		"tool_count", len(allTools),
	)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.provider.APIURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	switch c.provider.Type() {
	case "anthropic":
		httpReq.Header.Set("x-api-key", c.provider.APIKey())
		httpReq.Header.Set("anthropic-version", c.provider.APIVersion())
	case "openrouter":
		httpReq.Header.Set("Authorization", "Bearer "+c.provider.APIKey())
		httpReq.Header.Set("HTTP-Referer", "https://counterspell.dev")
	default:
		// Compatibility fallback for providers using Anthropic-compatible endpoint style.
		if strings.Contains(c.provider.APIURL(), "anthropic") {
			httpReq.Header.Set("x-api-key", c.provider.APIKey())
			httpReq.Header.Set("anthropic-version", c.provider.APIVersion())
		}
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
