package session

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/revrost/code/counterspell/pkg/orion"
	"github.com/lithammer/shortuuid/v4"
)

// Service provides in-memory session storage.
// It's suitable for testing and simple applications.
// For production use, consider implementing a
// persistent storage backend.
type Service struct {
	mu         sync.RWMutex
	sessions   map[string]orion.Session
	eventBroker orion.EventBroker[orion.Session]
}

// NewService creates a new in-memory session service.
func NewService(eventBroker orion.EventBroker[orion.Session]) orion.SessionService {
	return &Service{
		sessions:     make(map[string]orion.Session),
		eventBroker: eventBroker,
	}
}

// Create creates a new session with the given title.
func (s *Service) Create(ctx context.Context, title string) (orion.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := shortuuid.New()
	session := orion.NewSession(title)
	session.ID = id

	s.sessions[id] = session
	s.eventBroker.Publish("created", session)

	return session, nil
}

// CreateTaskSession creates a nested session for agent tool execution.
func (s *Service) CreateTaskSession(ctx context.Context, toolCallID, parentSessionID, title string) (orion.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := orion.NewSession(title)
	session.ID = toolCallID
	session.ParentSessionID = parentSessionID

	s.sessions[session.ID] = session
	s.eventBroker.Publish("created", session)

	return session, nil
}

// Get retrieves a session by ID.
func (s *Service) Get(ctx context.Context, id string) (orion.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	if !ok {
		return orion.Session{}, orion.ErrSessionNotFound
	}

	return session, nil
}

// List returns all sessions.
func (s *Service) List(ctx context.Context) ([]orion.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]orion.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Save persists changes to a session.
func (s *Service) Save(ctx context.Context, session orion.Session) (orion.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.sessions[session.ID]
	if !ok {
		return orion.Session{}, orion.ErrSessionNotFound
	}

	session.UpdatedAt = time.Now().Unix()
	s.sessions[session.ID] = session
	s.eventBroker.Publish("updated", session)

	return session, nil
}

// UpdateTitleAndUsage atomically updates title and usage fields.
func (s *Service) UpdateTitleAndUsage(ctx context.Context, sessionID, title string, promptTokens, completionTokens int64, cost float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return orion.ErrSessionNotFound
	}

	session.Title = title
	session.PromptTokens = promptTokens
	session.CompletionTokens = completionTokens
	session.Cost = cost
	session.UpdatedAt = time.Now().Unix()

	s.eventBroker.Publish("updated", session)
	return nil
}

// Delete removes a session.
func (s *Service) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return orion.ErrSessionNotFound
	}

	delete(s.sessions, id)
	s.eventBroker.Publish("deleted", session)

	return nil
}

// CreateAgentToolSessionID creates a session ID for agent tool sessions.
func (s *Service) CreateAgentToolSessionID(messageID, toolCallID string) string {
	return fmt.Sprintf("%s$$%s", messageID, toolCallID)
}

// ParseAgentToolSessionID parses an agent tool session ID into its components.
func (s *Service) ParseAgentToolSessionID(sessionID string) (messageID string, toolCallID string, ok bool) {
	parts := strings.Split(sessionID, "$$")
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// IsAgentToolSession checks if a session ID follows the agent tool session format.
func (s *Service) IsAgentToolSession(sessionID string) bool {
	_, _, ok := s.ParseAgentToolSessionID(sessionID)
	return ok
}

// Helper functions for todo marshaling (currently unused)

/*
func marshalTodos(todos []orion.Todo) (string, error) {
	if len(todos) == 0 {
		return "", nil
	}
	data, err := json.Marshal(todos)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshalTodos(data string) ([]orion.Todo, error) {
	if data == "" {
		return []orion.Todo{}, nil
	}
	var todos []orion.Todo
	if err := json.Unmarshal([]byte(data), &todos); err != nil {
		return []orion.Todo{}, err
	}
	return todos, nil
}
*/
