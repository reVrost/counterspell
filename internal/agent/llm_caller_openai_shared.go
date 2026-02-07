package agent

import (
	"encoding/json"
	"strings"

	"github.com/revrost/counterspell/internal/agent/tools"
)

// Shared OpenAI-compatible message/tool structures used by OpenAI-like callers.
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
