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

	"github.com/golang-jwt/jwt/v5"
	"github.com/revrost/counterspell/internal/agent/tools"
	"github.com/revrost/counterspell/internal/llm"
)

const (
	openAICodexResponsesURL = "https://chatgpt.com/backend-api/codex/responses"
)

// OpenAICaller implements LLMCaller for OpenAI connector-based chat.
type OpenAICaller struct {
	provider llm.Provider
}

type openAICodexRequest struct {
	Model        string `json:"model"`
	Instructions string `json:"instructions,omitempty"`
	Input        string `json:"input"`
	Store        bool   `json:"store"`
	Stream       bool   `json:"stream"`
}

func (c *OpenAICaller) Stream(ctx context.Context, messages []Message, allTools map[string]tools.Tool, systemPrompt string) (*LLMStream, error) {
	_ = allTools

	accessToken := strings.TrimSpace(c.provider.APIKey())
	if accessToken == "" {
		return nil, fmt.Errorf("openai connector is not connected")
	}
	accountID, err := openAIConnectorAccountID(accessToken)
	if err != nil {
		return nil, fmt.Errorf("openai connector is required: %w", err)
	}

	req := openAICodexRequest{
		Model:        c.provider.Model(),
		Instructions: strings.TrimSpace(systemPrompt),
		Input:        buildOpenAIConnectorInput(messages),
		Store:        false,
		Stream:       true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, openAICodexResponsesURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("chatgpt-account-id", accountID)
	httpReq.Header.Set("OpenAI-Beta", "responses=experimental")

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
		completed := false

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

			var payload map[string]any
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				continue
			}

			eventType := getStringValue(payload["type"])
			switch eventType {
			case "response.output_text.delta":
				delta := getStringValue(payload["delta"])
				if delta == "" {
					continue
				}
				if !textActive {
					if !emit(LLMEvent{Type: LLMContentStart, BlockType: "text", Block: &ContentBlock{Type: "text"}}) {
						done <- ctx.Err()
						return
					}
					textActive = true
				}
				if !emit(LLMEvent{Type: LLMContentDelta, BlockType: "text", Delta: delta}) {
					done <- ctx.Err()
					return
				}
			case "response.output_text.done":
				if textActive {
					if !emit(LLMEvent{Type: LLMContentEnd, BlockType: "text"}) {
						done <- ctx.Err()
						return
					}
					textActive = false
				}
			case "response.completed":
				completed = true
			case "response.failed", "error":
				done <- fmt.Errorf("openai connector response failed")
				return
			}

			if completed {
				break
			}
		}

		if err := scanner.Err(); err != nil {
			done <- err
			return
		}
		if textActive {
			_ = emit(LLMEvent{Type: LLMContentEnd, BlockType: "text"})
		}
		_ = emit(LLMEvent{Type: LLMMessageEnd})
		done <- nil
	}()

	return &LLMStream{Events: events, Done: done}, nil
}

func openAIConnectorAccountID(accessToken string) (string, error) {
	claims := jwt.MapClaims{}
	token, _, err := jwt.NewParser(jwt.WithoutClaimsValidation()).ParseUnverified(accessToken, claims)
	if err != nil || token == nil {
		return "", fmt.Errorf("invalid access token")
	}
	authClaimRaw, ok := claims["https://api.openai.com/auth"]
	if !ok {
		return "", fmt.Errorf("access token missing connector auth claims")
	}
	authClaim, ok := authClaimRaw.(map[string]any)
	if !ok {
		return "", fmt.Errorf("access token connector auth claims invalid")
	}
	accountID, _ := authClaim["chatgpt_account_id"].(string)
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return "", fmt.Errorf("access token missing chatgpt_account_id")
	}
	return accountID, nil
}

func buildOpenAIConnectorInput(messages []Message) string {
	var b strings.Builder
	for _, msg := range messages {
		var content strings.Builder
		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				if strings.TrimSpace(block.Text) != "" {
					if content.Len() > 0 {
						content.WriteString("\n")
					}
					content.WriteString(block.Text)
				}
			case "tool_result":
				if strings.TrimSpace(block.Content) != "" {
					if content.Len() > 0 {
						content.WriteString("\n")
					}
					content.WriteString(block.Content)
				}
			}
		}
		text := strings.TrimSpace(content.String())
		if text == "" {
			continue
		}
		fmt.Fprintf(&b, "%s: %s\n", strings.ToUpper(strings.TrimSpace(msg.Role)), text)
	}
	return strings.TrimSpace(b.String())
}

func getStringValue(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}
