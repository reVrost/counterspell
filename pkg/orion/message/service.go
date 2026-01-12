package message

import (
	"context"
	"time"
	"sync"

	"github.com/revrost/code/counterspell/pkg/orion"
	"github.com/lithammer/shortuuid/v4"
)

// Service provides in-memory message storage.
// It's suitable for testing and simple applications.
// For production use, consider implementing a
// persistent storage backend.
type Service struct {
	mu         sync.RWMutex
	messages   map[string]orion.Message
	eventBroker orion.EventBroker[orion.Message]
}

// NewService creates a new in-memory message service.
func NewService(eventBroker orion.EventBroker[orion.Message]) orion.MessageService {
	return &Service{
		messages:     make(map[string]orion.Message),
		eventBroker: eventBroker,
	}
}

// Create creates a new message in a session.
func (s *Service) Create(ctx context.Context, sessionID string, params orion.CreateMessageParams) (orion.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add finish part to non-assistant messages
	if params.Role != orion.RoleAssistant {
		params.Parts = append(params.Parts, orion.NewFinish(
			orion.FinishReasonEndTurn,
			"",
			"",
		))
	}

	id := shortuuid.New()
	msg := orion.NewMessage(sessionID, params.Role)
	msg.ID = id
	msg.Parts = params.Parts
	msg.Model = params.Model
	msg.Provider = params.Provider
	msg.IsSummaryMessage = params.IsSummaryMessage

	s.messages[id] = msg
	s.eventBroker.Publish("created", msg.Clone())

	return msg, nil
}

// Update updates an existing message.
func (s *Service) Update(ctx context.Context, message orion.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.messages[message.ID]
	if !ok {
		return orion.ErrMessageNotFound
	}

	message.UpdatedAt = time.Now().Unix()
	s.messages[message.ID] = message
	s.eventBroker.Publish("updated", message.Clone())

	return nil
}

// Get retrieves a message by ID.
func (s *Service) Get(ctx context.Context, id string) (orion.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg, ok := s.messages[id]
	if !ok {
		return orion.Message{}, orion.ErrMessageNotFound
	}

	return msg, nil
}

// List returns all messages in a session.
func (s *Service) List(ctx context.Context, sessionID string) ([]orion.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var messages []orion.Message
	for _, msg := range s.messages {
		if msg.SessionID == sessionID {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// Delete removes a message.
func (s *Service) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg, ok := s.messages[id]
	if !ok {
		return orion.ErrMessageNotFound
	}

	delete(s.messages, id)
	s.eventBroker.Publish("deleted", msg)

	return nil
}

// DeleteSessionMessages removes all messages in a session.
func (s *Service) DeleteSessionMessages(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var messages []orion.Message
	for _, msg := range s.messages {
		if msg.SessionID == sessionID {
			messages = append(messages, msg)
		}
	}

	for _, msg := range messages {
		delete(s.messages, msg.ID)
		s.eventBroker.Publish("deleted", msg)
	}

	return nil
}

// Message serialization helpers (commented out - currently unused)

/*
func marshallParts(parts []orion.ContentPart) ([]byte, error) {
	wrappedParts := make([]partWrapper, len(parts))
/*
	for i, part := range parts {
		wrappedParts[i] = partWrapper{
			Type: partTypeOf(part),
			Data: part,
		}
	}
	return json.Marshal(wrappedParts)
}

func unmarshallParts(data []byte) ([]orion.ContentPart, error) {
	temp := []json.RawMessage{}

	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}

	parts := make([]orion.ContentPart, 0)
	for _, rawPart := range temp {
		var wrapper struct {
			Type partType        `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(rawPart, &wrapper); err != nil {
			return nil, err
		}

		var part orion.ContentPart
		var err error

		switch wrapper.Type {
		case reasoningType:
			part, err = unmarshalReasoningContent(wrapper.Data)
		case textType:
			part, err = unmarshalTextContent(wrapper.Data)
		case toolCallType:
			part, err = unmarshalToolCall(wrapper.Data)
		case toolResultType:
			part, err = unmarshalToolResult(wrapper.Data)
		case finishType:
			part, err = unmarshalFinish(wrapper.Data)
		default:
			err = fmt.Errorf("unknown part type: %s", wrapper.Type)
		}

		if err != nil {
			return nil, err
		}

		parts = append(parts, part)
	}

	return parts, nil
}

func partTypeOf(part orion.ContentPart) partType {
	switch part.(type) {
	case orion.ReasoningContent:
		return reasoningType
	case orion.TextContent:
		return textType
	case orion.ToolCall:
		return toolCallType
	case orion.ToolResult:
		return toolResultType
	case orion.Finish:
		return finishType
	default:
		return ""
	}
}

func unmarshalReasoningContent(data json.RawMessage) (orion.ContentPart, error) {
	var rc orion.ReasoningContent
	if err := json.Unmarshal(data, &rc); err != nil {
		return nil, err
	}
	return rc, nil
}

func unmarshalTextContent(data json.RawMessage) (orion.ContentPart, error) {
	var tc orion.TextContent
	if err := json.Unmarshal(data, &tc); err != nil {
		return nil, err
	}
	return tc, nil
}

func unmarshalToolCall(data json.RawMessage) (orion.ContentPart, error) {
	var tc orion.ToolCall
	if err := json.Unmarshal(data, &tc); err != nil {
		return nil, err
	}
	return tc, nil
}

func unmarshalToolResult(data json.RawMessage) (orion.ContentPart, error) {
	var tr orion.ToolResult
	if err := json.Unmarshal(data, &tr); err != nil {
		return nil, err
	}
	return tr, nil
}

func unmarshalFinish(data json.RawMessage) (orion.ContentPart, error) {
	var f orion.Finish
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f, nil
}

*/
// type partType string
// 
// const (
// 	reasoningType  partType = "reasoning"
// 	textType       partType = "text"
// 	toolCallType   partType = "tool_call"
// 	toolResultType partType = "tool_result"
// 	finishType     partType = "finish"
// )
// 
// // partWrapper represents a serialized content part.
// type partWrapper struct {
// 	Type partType         `json:"type"`
// 	Data orion.ContentPart `json:"data"`
// }
