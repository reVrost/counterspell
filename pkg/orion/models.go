package orion

import (
	"fmt"
	"time"
)

// FinishReason constants define why an agent finished processing.
const (
	// FinishReasonEndTurn indicates normal completion.
	FinishReasonEndTurn = "end_turn"

	// FinishReasonToolUse indicates that agent used tools.
	FinishReasonToolUse = "tool_use"

	// FinishReasonMaxTokens indicates that output reached max tokens.
	FinishReasonMaxTokens = "max_tokens"

	// FinishReasonCanceled indicates that request was canceled.
	FinishReasonCanceled = "canceled"

	// FinishReasonPermissionDenied indicates that permission was denied.
	FinishReasonPermissionDenied = "permission_denied"

	// FinishReasonError indicates that an error occurred.
	FinishReasonError = "error"

	// FinishReasonUnknown indicates an unknown finish reason.
	FinishReasonUnknown = "unknown"
)

// Content returns text content from a message's parts.
func (m *Message) Content() TextContent {
	for _, part := range m.Parts {
		if tc, ok := part.(TextContent); ok {
			return tc
		}
	}
	return TextContent{}
}

// ReasoningContent returns reasoning content from a message's parts.
func (m *Message) ReasoningContent() ReasoningContent {
	for _, part := range m.Parts {
		if rc, ok := part.(ReasoningContent); ok {
			return rc
		}
	}
	return ReasoningContent{}
}

// ToolCalls returns all tool call parts from a message.
func (m *Message) ToolCalls() []ToolCall {
	var calls []ToolCall
	for _, part := range m.Parts {
		if tc, ok := part.(ToolCall); ok {
			calls = append(calls, tc)
		}
	}
	return calls
}

// ToolResults returns all tool result parts from a message.
func (m *Message) ToolResults() []ToolResult {
	var results []ToolResult
	for _, part := range m.Parts {
		if tr, ok := part.(ToolResult); ok {
			results = append(results, tr)
		}
	}
	return results
}

// FinishPart returns the finish part from a message.
func (m *Message) FinishPart() *Finish {
	for _, part := range m.Parts {
		if f, ok := part.(Finish); ok {
			return &f
		}
	}
	return nil
}

// Clone returns a deep copy of the message.
func (m *Message) Clone() Message {
	clone := *m

	// Clone parts slice
	clone.Parts = make([]ContentPart, len(m.Parts))
	for i, part := range m.Parts {
		clone.Parts[i] = cloneContentPart(part)
	}

	return clone
}

// cloneContentPart creates a deep copy of a content part.
func cloneContentPart(part ContentPart) ContentPart {
	switch p := part.(type) {
	case TextContent:
		return TextContent{Text: p.Text}
	case ReasoningContent:
		return ReasoningContent{Text: p.Text, Signature: p.Signature}
	case ToolCall:
		return ToolCall{
			ID:               p.ID,
			Name:             p.Name,
			Input:            p.Input,
			ProviderExecuted: p.ProviderExecuted,
			Finished:         p.Finished,
		}
	case ToolResult:
		metadata := make(map[string]interface{})
		for k, v := range p.Metadata {
			metadata[k] = v
		}
		return ToolResult{
			ToolCallID: p.ToolCallID,
			Name:       p.Name,
			Content:    p.Content,
			IsError:    p.IsError,
			Metadata:   metadata,
		}
	case Finish:
		return Finish{
			Reason:      p.Reason,
			Title:       p.Title,
			Description: p.Description,
			Time:        p.Time,
		}
	default:
		panic(fmt.Sprintf("unknown content part type: %T", part))
	}
}

// AppendContent adds text content to the message.
func (m *Message) AppendContent(text string) {
	// Find existing text content part or create a new one
	for i, part := range m.Parts {
		if tc, ok := part.(TextContent); ok {
			tc.Text += text
			m.Parts[i] = tc
			return
		}
	}
	// No text content found, append a new one
	m.Parts = append(m.Parts, TextContent{Text: text})
}

// AppendReasoningContent adds reasoning content to the message.
func (m *Message) AppendReasoningContent(text string) {
	// Find existing reasoning content part or create a new one
	for i, part := range m.Parts {
		if rc, ok := part.(ReasoningContent); ok {
			rc.Text += text
			m.Parts[i] = rc
			return
		}
	}
	// No reasoning content found, append a new one
	m.Parts = append(m.Parts, ReasoningContent{Text: text})
}

// AppendReasoningSignature adds a signature to the reasoning content.
func (m *Message) AppendReasoningSignature(signature string) {
	for i, part := range m.Parts {
		if rc, ok := part.(ReasoningContent); ok {
			rc.Signature = signature
			m.Parts[i] = rc
			return
		}
	}
}

// AddToolCall adds a tool call to the message.
func (m *Message) AddToolCall(call ToolCall) {
	m.Parts = append(m.Parts, call)
}

// AddFinish adds a finish part to the message.
func (m *Message) AddFinish(reason, title, description string) {
	// Remove existing finish part if present
	var newParts []ContentPart
	for _, part := range m.Parts {
		if _, ok := part.(Finish); !ok {
			newParts = append(newParts, part)
		}
	}
	newParts = append(newParts, Finish{
		Reason:      reason,
		Title:       title,
		Description: description,
	})
	m.Parts = newParts
}

// FinishThinking marks reasoning content as complete.
func (m *Message) FinishThinking() {
	// No-op - this is for compatibility with Crush's API
}

// Clone returns a deep copy of the session.
func (s *Session) Clone() Session {
	clone := *s

	// Clone todos slice
	clone.Todos = make([]Todo, len(s.Todos))
	for i, todo := range s.Todos {
		clone.Todos[i] = Todo{
			Content:    todo.Content,
			Status:     todo.Status,
			ActiveForm: todo.ActiveForm,
		}
	}

	return clone
}

// AttachmentIsText checks if an attachment is text content.
func AttachmentIsText(a Attachment) bool {
	return a.IsTextFile
}

// NewSession creates a new session with the given title.
func NewSession(title string) Session {
	now := time.Now().Unix()
	return Session{
		ID:               "",
		ParentSessionID:  "",
		Title:            title,
		MessageCount:     0,
		PromptTokens:     0,
		CompletionTokens: 0,
		SummaryMessageID: "",
		Cost:             0.0,
		Todos:            []Todo{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewMessage creates a new message with the given parameters.
func NewMessage(sessionID string, role MessageRole) Message {
	now := time.Now().Unix()
	return Message{
		ID:               "",
		SessionID:        sessionID,
		Role:             role,
		Parts:            []ContentPart{},
		Model:            "",
		Provider:         "",
		IsSummaryMessage: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// AddFinish updates the message with a finish reason.
func AddFinishToMessage(msg *Message, reason, title, description string) {
	msg.AddFinish(reason, title, description)
	// Set the finish time if available
	for i, part := range msg.Parts {
		if f, ok := part.(Finish); ok {
			f.Time = time.Now().Unix()
			msg.Parts[i] = f
			break
		}
	}
}
