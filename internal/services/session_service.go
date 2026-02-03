package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"github.com/revrost/counterspell/internal/agent"
	"github.com/revrost/counterspell/internal/llm"
	"github.com/revrost/counterspell/internal/models"
)

var ErrCodexUnsupported = errors.New("codex sessions are not supported yet")

// SessionService manages session CRUD and chat execution.
type SessionService struct {
	repo     *Repository
	settings *SettingsService
	dataDir  string
}

// NewSessionService creates a new SessionService.
func NewSessionService(repo *Repository, settings *SettingsService, dataDir string) *SessionService {
	return &SessionService{
		repo:     repo,
		settings: settings,
		dataDir:  dataDir,
	}
}

// List returns all sessions.
func (s *SessionService) List(ctx context.Context) ([]*models.Session, error) {
	return s.repo.ListSessions(ctx)
}

// Get returns a session with messages.
func (s *SessionService) Get(ctx context.Context, sessionID string) (*models.Session, []models.SessionMessage, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}
	messages, err := s.repo.ListSessionMessages(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}
	if session.AgentBackend == "codex" {
		messages = filterCodexSetupMessages(messages)
	}
	return session, messages, nil
}

// Create creates a new empty session.
func (s *SessionService) Create(ctx context.Context, backend string) (*models.Session, error) {
	backend = strings.TrimSpace(backend)
	if backend == "" {
		settings, err := s.settings.GetSettings(ctx)
		if err == nil && settings != nil && settings.AgentBackend != "" {
			backend = settings.AgentBackend
		} else {
			backend = "native"
		}
	}
	if backend == "codex" {
		return nil, ErrCodexUnsupported
	}

	now := time.Now().UnixMilli()
	id := shortuuid.New()
	externalID := id
	title := "New session"

	session := &models.Session{
		ID:           id,
		AgentBackend: backend,
		ExternalID:   &externalID,
		Title:        &title,
		MessageCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return s.repo.CreateSession(ctx, session)
}

// Chat sends a message to a session and stores responses.
func (s *SessionService) Chat(ctx context.Context, sessionID, message, modelID string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("message is required")
	}

	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session.AgentBackend == "codex" {
		return ErrCodexUnsupported
	}

	existingMessages, err := s.repo.ListSessionMessages(ctx, sessionID)
	if err != nil {
		return err
	}

	nextSeq, err := s.repo.GetSessionNextSequence(ctx, sessionID)
	if err != nil {
		return err
	}

	now := time.Now().UnixMilli()
	userRaw, _ := json.Marshal(map[string]any{
		"role":    "user",
		"kind":    "text",
		"content": message,
	})
	if err := s.repo.CreateSessionMessage(
		ctx,
		sessionID,
		nextSeq,
		"user",
		"text",
		message,
		"",
		"",
		string(userRaw),
		now,
	); err != nil {
		return err
	}

	if session.Title == nil || strings.TrimSpace(*session.Title) == "" || *session.Title == "New session" {
		title := truncateSessionTitle(message)
		_ = s.repo.UpdateSessionTitle(ctx, sessionID, title)
	}

	writer := newSessionMessageWriter(ctx, s.repo, sessionID, nextSeq+1)
	backend, cleanup, err := s.buildBackend(ctx, session, modelID, s.makeSessionCallback(writer), true)
	if err != nil {
		return err
	}
	defer cleanup()

	if session.AgentBackend == "native" {
		historyJSON, err := buildNativeHistory(existingMessages)
		if err != nil {
			return err
		}
		if err := backend.RestoreState(historyJSON); err != nil {
			return err
		}
	}

	if err := backend.Run(ctx, message); err != nil {
		return err
	}

	return nil
}

// Promote converts a session into a task with summarized title/intent.
func (s *SessionService) Promote(ctx context.Context, sessionID string) (*models.Task, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetTaskBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	messages, err := s.repo.ListSessionMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.AgentBackend == "codex" {
		return nil, ErrCodexUnsupported
	}

	snapshot, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to snapshot messages: %w", err)
	}

	summaryTitle, summaryIntent, err := s.summarizeSession(ctx, session, messages)
	if err != nil {
		slog.Warn("[SESSIONS] summarize failed", "session_id", sessionID, "error", err)
		return nil, fmt.Errorf("failed to summarize session: %w", err)
	}
	if summaryTitle == "" || summaryIntent == "" {
		return nil, fmt.Errorf("failed to summarize session: empty title or intent")
	}

	task, err := s.repo.CreateFromSession(ctx, sessionID, summaryTitle, summaryIntent, string(snapshot))
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *SessionService) summarizeSession(ctx context.Context, session *models.Session, messages []models.SessionMessage) (string, string, error) {
	prompt := buildSummaryPrompt(messages)
	backend, cleanup, err := s.buildBackend(ctx, session, "", nil, false)
	if err != nil {
		return "", "", err
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if err := backend.Run(ctx, prompt); err != nil {
		return "", "", err
	}

	raw := strings.TrimSpace(backend.FinalMessage())
	if raw == "" {
		return "", "", fmt.Errorf("empty summary response")
	}

	title, intent := parseSummaryResponse(raw)
	if title == "" || intent == "" {
		return "", "", fmt.Errorf("unable to parse summary response")
	}
	return title, intent, nil
}

func (s *SessionService) buildBackend(
	ctx context.Context,
	session *models.Session,
	modelID string,
	callback agent.StreamCallback,
	useSessionID bool,
) (agent.Backend, func(), error) {
	if session.AgentBackend == "codex" {
		return nil, func() {}, ErrCodexUnsupported
	}

	apiKey, provider, model, err := s.resolveProvider(ctx, modelID)
	if err != nil {
		return nil, func() {}, err
	}

	switch session.AgentBackend {
	case "claude-code":
		baseURL := ""
		switch provider {
		case "zai":
			baseURL = "https://api.z.ai/api/anthropic"
		case "openrouter":
			baseURL = "https://openrouter.ai/api"
		}

		opts := []agent.ClaudeCodeOption{
			agent.WithAPIKey(apiKey),
			agent.WithModel(model),
			agent.WithBaseURL(baseURL),
			agent.WithClaudeWorkDir(s.dataDir),
		}
		if callback != nil {
			opts = append(opts, agent.WithClaudeCallback(callback))
		}
		if useSessionID && session.BackendSessionID != nil && *session.BackendSessionID != "" {
			opts = append(opts, agent.WithSessionID(*session.BackendSessionID))
		}
		backend, err := agent.NewClaudeCodeBackend(opts...)
		if err != nil {
			return nil, func() {}, err
		}
		return backend, func() { _ = backend.Close() }, nil
	case "native":
		llmProvider, err := newLLMProvider(provider, apiKey)
		if err != nil {
			return nil, func() {}, err
		}
		llmProvider.SetModel(model)
		opts := []agent.NativeBackendOption{
			agent.WithProvider(llmProvider),
			agent.WithWorkDir(s.dataDir),
		}
		if callback != nil {
			opts = append(opts, agent.WithCallback(callback))
		}
		backend, err := agent.NewNativeBackend(opts...)
		if err != nil {
			return nil, func() {}, err
		}
		return backend, func() { _ = backend.Close() }, nil
	default:
		return nil, func() {}, fmt.Errorf("unsupported agent_backend: %s", session.AgentBackend)
	}
}

func (s *SessionService) resolveProvider(ctx context.Context, modelID string) (string, string, string, error) {
	provider := ""
	model := ""

	if modelID != "" {
		parts := strings.SplitN(modelID, "#", 2)
		if len(parts) == 2 {
			provider = parts[0]
			model = parts[1]
			switch provider {
			case "o":
				provider = "openrouter"
			case "zai":
				provider = "zai"
			}
		} else {
			model = parts[0]
		}
	}

	apiKey, actualProvider, actualModel, err := s.settings.GetAPIKeyForProvider(ctx, provider)
	if err != nil {
		return "", "", "", err
	}
	if model == "" {
		model = actualModel
	}
	return apiKey, actualProvider, model, nil
}

func newLLMProvider(provider, apiKey string) (llm.Provider, error) {
	switch provider {
	case "anthropic":
		return llm.NewAnthropicProvider(apiKey), nil
	case "openrouter":
		return llm.NewOpenRouterProvider(apiKey), nil
	case "zai":
		return llm.NewZaiProvider(apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func buildNativeHistory(messages []models.SessionMessage) (string, error) {
	history := make([]agent.Message, 0, len(messages))
	for _, msg := range messages {
		content := ""
		if msg.Content != nil {
			content = *msg.Content
		}
		switch msg.Kind {
		case "tool_use":
			input := map[string]any{}
			if content != "" {
				_ = json.Unmarshal([]byte(content), &input)
			}
			history = append(history, agent.Message{
				Role: "assistant",
				Content: []agent.ContentBlock{{
					Type:  "tool_use",
					Name:  valueOrEmpty(msg.ToolName),
					ID:    valueOrEmpty(msg.ToolCallID),
					Input: input,
				}},
			})
		case "tool_result":
			history = append(history, agent.Message{
				Role: "user",
				Content: []agent.ContentBlock{{
					Type:      "tool_result",
					ToolUseID: valueOrEmpty(msg.ToolCallID),
					Content:   content,
				}},
			})
		default:
			if strings.TrimSpace(content) == "" {
				continue
			}
			if msg.Role != "user" && msg.Role != "assistant" {
				continue
			}
			history = append(history, agent.Message{
				Role: msg.Role,
				Content: []agent.ContentBlock{{
					Type: "text",
					Text: content,
				}},
			})
		}
	}

	data, err := json.Marshal(history)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildSummaryPrompt(messages []models.SessionMessage) string {
	var b strings.Builder
	b.WriteString("Summarize this session into a task. Return ONLY valid JSON with keys \"title\" and \"intent\".\n")
	b.WriteString("Title: short 5-12 words. Intent: full, detailed description with requirements and constraints.\n\n")
	b.WriteString("Session transcript:\n")

	for _, msg := range messages {
		role := strings.ToUpper(msg.Role)
		kind := msg.Kind
		content := ""
		if msg.Content != nil {
			content = *msg.Content
		}
		if content == "" && kind == "text" {
			continue
		}
		switch kind {
		case "tool_use":
			fmt.Fprintf(&b, "[%s TOOL_USE %s] %s\n", role, valueOrEmpty(msg.ToolName), content)
		case "tool_result":
			fmt.Fprintf(&b, "[%s TOOL_RESULT] %s\n", role, content)
		default:
			fmt.Fprintf(&b, "[%s] %s\n", role, content)
		}
	}

	return b.String()
}

func parseSummaryResponse(raw string) (string, string) {
	type summary struct {
		Title  string `json:"title"`
		Intent string `json:"intent"`
	}

	var parsed summary
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		return strings.TrimSpace(parsed.Title), strings.TrimSpace(parsed.Intent)
	}

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start != -1 && end != -1 && end > start {
		if err := json.Unmarshal([]byte(raw[start:end+1]), &parsed); err == nil {
			return strings.TrimSpace(parsed.Title), strings.TrimSpace(parsed.Intent)
		}
	}

	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return "", ""
	}
	title := strings.TrimSpace(lines[0])
	intent := strings.TrimSpace(strings.Join(lines[1:], "\n"))
	return title, intent
}

func truncateSessionTitle(title string) string {
	const maxLen = 120
	trimmed := strings.TrimSpace(title)
	if len(trimmed) <= maxLen {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:maxLen]) + "..."
}

type sessionMessageWriter struct {
	ctx       context.Context
	repo      *Repository
	sessionID string
	nextSeq   int64
	nextSeqMu sync.Mutex
}

func newSessionMessageWriter(ctx context.Context, repo *Repository, sessionID string, startSeq int64) *sessionMessageWriter {
	return &sessionMessageWriter{
		ctx:       ctx,
		repo:      repo,
		sessionID: sessionID,
		nextSeq:   startSeq,
	}
}

func (w *sessionMessageWriter) nextSequence() int64 {
	w.nextSeqMu.Lock()
	defer w.nextSeqMu.Unlock()
	seq := w.nextSeq
	w.nextSeq++
	return seq
}

func (w *sessionMessageWriter) append(role, kind, content, toolName, toolCallID string, raw map[string]any) {
	if kind == "" {
		kind = "text"
	}
	seq := w.nextSequence()
	now := time.Now().UnixMilli()
	rawJSON, _ := json.Marshal(raw)

	if err := w.repo.CreateSessionMessage(
		w.ctx,
		w.sessionID,
		seq,
		role,
		kind,
		content,
		toolName,
		toolCallID,
		string(rawJSON),
		now,
	); err != nil {
		slog.Warn("[SESSIONS] failed to persist session message", "session_id", w.sessionID, "error", err)
	}
}

func (s *SessionService) makeSessionCallback(writer *sessionMessageWriter) agent.StreamCallback {
	return func(event agent.StreamEvent) {
		switch event.Type {
		case "session":
			if event.SessionID != "" {
				if err := s.repo.UpdateSessionBackendSessionID(context.Background(), writer.sessionID, event.SessionID); err != nil {
					slog.Warn("[SESSIONS] failed to update backend session id", "session_id", writer.sessionID, "error", err)
				}
			}
		case agent.EventText:
			writer.append(
				"assistant",
				"text",
				event.Content,
				"",
				"",
				map[string]any{"type": "text", "content": event.Content},
			)
		case agent.EventTool:
			writer.append(
				"assistant",
				"tool_use",
				event.Args,
				event.Tool,
				"",
				map[string]any{"type": "tool_use", "tool": event.Tool, "args": event.Args},
			)
		case agent.EventToolResult:
			writer.append(
				"tool",
				"tool_result",
				event.Content,
				"",
				"",
				map[string]any{"type": "tool_result", "content": event.Content},
			)
		default:
		}
	}
}
