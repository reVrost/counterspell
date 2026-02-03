package services

import (
	"encoding/json"
	"strings"

	"github.com/revrost/counterspell/internal/agent"
)

type streamMessage struct {
	id      string
	role    string
	blocks  []agent.ContentBlock
	current *agent.ContentBlock
	argsBuf strings.Builder
}

type streamAssembler struct {
	messages map[string]*streamMessage
}

func newStreamAssembler() *streamAssembler {
	return &streamAssembler{messages: make(map[string]*streamMessage)}
}

func (a *streamAssembler) ensureMessage(id, role string) *streamMessage {
	msg := a.messages[id]
	if msg == nil {
		msg = &streamMessage{id: id, role: role}
		a.messages[id] = msg
	}
	if msg.role == "" {
		msg.role = role
	}
	return msg
}

func (a *streamAssembler) Apply(event agent.StreamEvent) (*streamMessage, bool) {
	switch event.Type {
	case agent.EventMessageStart:
		if event.MessageID == "" {
			return nil, false
		}
		a.ensureMessage(event.MessageID, event.Role)
	case agent.EventContentStart:
		if event.MessageID == "" {
			return nil, false
		}
		msg := a.ensureMessage(event.MessageID, event.Role)
		block := event.Block
		if block == nil {
			block = &agent.ContentBlock{Type: event.BlockType}
		}
		msg.current = block
		msg.argsBuf.Reset()
	case agent.EventContentDelta:
		if event.MessageID == "" || event.BlockType == "" {
			return nil, false
		}
		msg := a.ensureMessage(event.MessageID, event.Role)
		if msg.current == nil || msg.current.Type != event.BlockType {
			msg.current = &agent.ContentBlock{Type: event.BlockType}
			msg.argsBuf.Reset()
		}
		switch event.BlockType {
		case "text", "thinking":
			msg.current.Text += event.Delta
		case "tool_use":
			msg.argsBuf.WriteString(event.Delta)
		}
	case agent.EventContentEnd:
		if event.MessageID == "" {
			return nil, false
		}
		msg := a.ensureMessage(event.MessageID, event.Role)
		block := event.Block
		if block == nil {
			block = msg.current
		}
		if block == nil {
			return nil, false
		}
		if block.Type == "tool_use" && block.Input == nil {
			raw := strings.TrimSpace(msg.argsBuf.String())
			if raw != "" {
				input := map[string]any{}
				if err := json.Unmarshal([]byte(raw), &input); err == nil {
					block.Input = input
				} else {
					block.Input = map[string]any{"raw": raw}
				}
			}
		}
		msg.blocks = append(msg.blocks, *block)
		msg.current = nil
		msg.argsBuf.Reset()
	case agent.EventMessageEnd:
		if event.MessageID == "" {
			return nil, false
		}
		msg := a.messages[event.MessageID]
		if msg == nil {
			return nil, false
		}
		delete(a.messages, event.MessageID)
		return msg, true
	}
	return nil, false
}
