package services

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/revrost/counterspell/internal/models"
)

const (
	defaultSessionSyncInterval = 5 * time.Second
	backendClaudeCode          = "claude-code"
	backendCodex               = "codex"
)

type importedMessage struct {
	Role    string
	Content string
}

type SessionSyncer struct {
	repo   *Repository
	events *EventBus

	pollInterval time.Duration
	stopCh       chan struct{}

	lastSeenMu sync.Mutex
	lastSeen   map[string]time.Time

	scanMu sync.Mutex
}

func NewSessionSyncer(repo *Repository, events *EventBus) *SessionSyncer {
	return &SessionSyncer{
		repo:         repo,
		events:       events,
		pollInterval: defaultSessionSyncInterval,
		stopCh:       make(chan struct{}),
		lastSeen:     make(map[string]time.Time),
	}
}

func (s *SessionSyncer) Start(ctx context.Context) {
	slog.Info("[SESSION-SYNC] starting", "interval", s.pollInterval.String())
	s.scan(ctx, true)

	go func() {
		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("[SESSION-SYNC] context cancelled, stopping")
				return
			case <-s.stopCh:
				slog.Info("[SESSION-SYNC] stop requested")
				return
			case <-ticker.C:
				s.scan(ctx, false)
			}
		}
	}()
}

func (s *SessionSyncer) Shutdown() {
	close(s.stopCh)
}

func (s *SessionSyncer) scan(ctx context.Context, force bool) {
	s.scanMu.Lock()
	defer s.scanMu.Unlock()

	claudeRoot := envPath("COUNTERSPELL_CLAUDE_DIR", filepath.Join(userHomeDir(), ".claude", "projects"))
	codexRoot := envPath("COUNTERSPELL_CODEX_DIR", filepath.Join(userHomeDir(), ".codex", "sessions"))

	claudeFiles, err := discoverClaudeTranscripts(claudeRoot)
	if err != nil {
		slog.Warn("[SESSION-SYNC] claude discovery failed", "error", err)
	} else {
		s.importFiles(ctx, backendClaudeCode, claudeFiles, force)
	}

	codexFiles, err := discoverCodexSessions(codexRoot)
	if err != nil {
		slog.Warn("[SESSION-SYNC] codex discovery failed", "error", err)
	} else {
		s.importFiles(ctx, backendCodex, codexFiles, force)
	}
}

func (s *SessionSyncer) importFiles(ctx context.Context, backend string, files []string, force bool) {
	for _, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if !force && !s.isFileUpdated(path, info.ModTime()) {
			continue
		}

		var sessionID string
		var messages []importedMessage
		switch backend {
		case backendClaudeCode:
			sessionID, messages, err = parseClaudeTranscript(path)
		case backendCodex:
			sessionID, messages, err = parseCodexSession(path)
		default:
			continue
		}
		if err != nil {
			slog.Warn("[SESSION-SYNC] failed to parse session", "backend", backend, "path", path, "error", err)
			continue
		}

		sessionID = normalizeSessionID(sessionID, path)
		if sessionID == "" || len(messages) == 0 {
			s.markFileSeen(path, info.ModTime())
			continue
		}

		if err := s.syncSession(ctx, backend, sessionID, messages); err != nil {
			slog.Warn("[SESSION-SYNC] failed to sync session", "backend", backend, "session_id", sessionID, "error", err)
			continue
		}

		s.markFileSeen(path, info.ModTime())
	}
}

func (s *SessionSyncer) isFileUpdated(path string, modTime time.Time) bool {
	s.lastSeenMu.Lock()
	defer s.lastSeenMu.Unlock()

	last, ok := s.lastSeen[path]
	if !ok {
		return true
	}
	return modTime.After(last)
}

func (s *SessionSyncer) markFileSeen(path string, modTime time.Time) {
	s.lastSeenMu.Lock()
	defer s.lastSeenMu.Unlock()
	s.lastSeen[path] = modTime
}

func (s *SessionSyncer) syncSession(ctx context.Context, backend, sessionID string, messages []importedMessage) error {
	run, err := s.repo.GetAgentRunByBackendSessionID(ctx, backend, sessionID)
	if err != nil {
		return err
	}

	created := false
	if run == nil {
		intent := sessionIntent(messages)
		task, err := s.repo.Create(ctx, "", intent)
		if err != nil {
			return err
		}

		runID, err := s.repo.CreateAgentRun(ctx, task.ID, intent, backend, "", "")
		if err != nil {
			return err
		}

		if err := s.repo.UpdateAgentRunBackendSessionID(ctx, runID, sessionID); err != nil {
			return err
		}

		run, err = s.repo.GetAgentRun(ctx, runID)
		if err != nil {
			return err
		}

		created = true
	}

	start := int(run.MessageCount)
	if start < 0 {
		start = 0
	}
	if start >= len(messages) {
		return nil
	}

	for _, msg := range messages[start:] {
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}
		if err := s.repo.CreateMessage(ctx, run.TaskID, run.ID, msg.Role, msg.Content); err != nil {
			return err
		}
	}

	s.publishUpdates(run.TaskID, created)
	return nil
}

func (s *SessionSyncer) publishUpdates(taskID string, created bool) {
	eventType := EventTypeTaskUpdated
	if created {
		eventType = EventTypeTaskStarted
	}

	s.events.Publish(models.Event{
		TaskID:    taskID,
		Type:      string(eventType),
		Data:      "",
		CreatedAt: time.Now(),
	})

	s.events.Publish(models.Event{
		TaskID:    taskID,
		Type:      string(EventTypeAgentRunUpdated),
		Data:      "",
		CreatedAt: time.Now(),
	})
}

func discoverClaudeTranscripts(root string) ([]string, error) {
	if root == "" {
		return nil, nil
	}
	if _, err := os.Stat(root); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".jsonl") {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"transcripts"+string(filepath.Separator)) {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

func discoverCodexSessions(root string) ([]string, error) {
	if root == "" {
		return nil, nil
	}
	if _, err := os.Stat(root); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext == ".json" || ext == ".jsonl" {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

func parseClaudeTranscript(path string) (string, []importedMessage, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	var sessionID string
	var messages []importedMessage

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		if sessionID == "" {
			if value, ok := event["session_id"].(string); ok && value != "" {
				sessionID = value
			}
		}

		eventType, _ := event["type"].(string)
		switch eventType {
		case "system":
			if sessionID == "" {
				if value, ok := event["session_id"].(string); ok && value != "" {
					sessionID = value
				}
			}
		case "user", "assistant":
			role := eventType
			content := ""
			if message, ok := event["message"].(map[string]any); ok {
				content = extractTextFromContent(message["content"])
			}
			if content == "" {
				content = extractTextFromContent(event["content"])
			}
			content = strings.TrimSpace(content)
			if content != "" {
				messages = append(messages, importedMessage{Role: role, Content: content})
			}
		case "user_text":
			content := strings.TrimSpace(extractTextFromContent(event["content"]))
			if content != "" {
				messages = append(messages, importedMessage{Role: "user", Content: content})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}

	return sessionID, messages, nil
}

func parseCodexSession(path string) (string, []importedMessage, error) {
	sessionID, messages, err := parseCodexJSONL(path)
	if err != nil {
		return "", nil, err
	}
	if len(messages) > 0 {
		return sessionID, messages, nil
	}

	if strings.ToLower(filepath.Ext(path)) == ".jsonl" {
		return sessionID, messages, nil
	}

	return parseCodexJSON(path)
}

func parseCodexJSONL(path string) (string, []importedMessage, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	var sessionID string
	var messages []importedMessage

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			continue
		}
		if sessionID == "" {
			sessionID = extractCodexSessionID(payload)
		}
		if msg, ok := extractCodexMessage(payload); ok {
			messages = append(messages, msg)
			continue
		}
		if msg, ok := extractMessage(payload); ok {
			messages = append(messages, msg)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}

	return sessionID, messages, nil
}

func parseCodexJSON(path string) (string, []importedMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", nil, err
	}

	sessionID := extractSessionID(payload)
	rawMessages, ok := payload["messages"].([]any)
	if !ok {
		if conversation, ok := payload["conversation"].([]any); ok {
			rawMessages = conversation
		}
	}

	messages := make([]importedMessage, 0)
	for _, raw := range rawMessages {
		if msgMap, ok := raw.(map[string]any); ok {
			if msg, ok := extractMessage(msgMap); ok {
				messages = append(messages, msg)
			}
		}
	}

	return sessionID, messages, nil
}


func extractCodexSessionID(event map[string]any) string {
	if eventType, ok := event["type"].(string); ok && eventType == "session_meta" {
		if payload, ok := event["payload"].(map[string]any); ok {
			if id, ok := payload["id"].(string); ok && id != "" {
				return id
			}
		}
	}
	return extractSessionID(event)
}

func extractCodexMessage(event map[string]any) (importedMessage, bool) {
	eventType, _ := event["type"].(string)
	if eventType != "response_item" {
		return importedMessage{}, false
	}

	payload, ok := event["payload"].(map[string]any)
	if !ok {
		return importedMessage{}, false
	}

	itemType, _ := payload["type"].(string)
	if itemType != "message" {
		return importedMessage{}, false
	}

	role, _ := payload["role"].(string)
	if role == "" {
		role = "assistant"
	}

	content := extractTextFromContent(payload["content"])
	content = strings.TrimSpace(content)
	if content == "" {
		return importedMessage{}, false
	}

	return importedMessage{Role: role, Content: content}, true
}

func extractMessage(payload map[string]any) (importedMessage, bool) {
	if msg, ok := payload["message"].(map[string]any); ok {
		return extractMessage(msg)
	}

	role, _ := payload["role"].(string)
	if role == "" {
		if msgType, ok := payload["type"].(string); ok {
			switch msgType {
			case "user", "assistant", "tool", "system":
				role = msgType
			}
		}
	}
	if role == "" {
		return importedMessage{}, false
	}

	content := extractTextFromContent(payload["content"])
	if content == "" {
		content = extractTextFromContent(payload["text"])
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return importedMessage{}, false
	}

	return importedMessage{Role: role, Content: content}, true
}

func extractSessionID(payload map[string]any) string {
	if val, ok := payload["session_id"].(string); ok && val != "" {
		return val
	}
	if val, ok := payload["sessionId"].(string); ok && val != "" {
		return val
	}
	if val, ok := payload["id"].(string); ok && val != "" {
		return val
	}
	return ""
}

func extractTextFromContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			switch block := item.(type) {
			case string:
				parts = append(parts, block)
			case map[string]any:
				if text := extractTextFromBlock(block); text != "" {
					parts = append(parts, text)
				}
			}
		}
		return strings.Join(parts, "")
	case map[string]any:
		if text := extractTextFromBlock(v); text != "" {
			return text
		}
		if inner, ok := v["content"]; ok {
			return extractTextFromContent(inner)
		}
	}
	return ""
}

func extractTextFromBlock(block map[string]any) string {
	if block == nil {
		return ""
	}

	if blockType, ok := block["type"].(string); ok {
		switch blockType {
		case "text", "output_text", "input_text":
			if text, ok := block["text"].(string); ok {
				return text
			}
		default:
			return ""
		}
	}

	if text, ok := block["text"].(string); ok {
		return text
	}
	if text, ok := block["content"].(string); ok {
		return text
	}
	return ""
}

func normalizeSessionID(sessionID, path string) string {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID != "" {
		return sessionID
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if base != "" {
		return base
	}
	return hashPath(path)
}

func sessionIntent(messages []importedMessage) string {
	for _, msg := range messages {
		if msg.Role == "user" && strings.TrimSpace(msg.Content) != "" {
			return msg.Content
		}
	}
	for _, msg := range messages {
		if strings.TrimSpace(msg.Content) != "" {
			return msg.Content
		}
	}
	return "Imported session"
}

func envPath(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func hashPath(path string) string {
	if path == "" {
		return ""
	}
	sum := sha1.Sum([]byte(path))
	return "path-" + hex.EncodeToString(sum[:8])
}
