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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"github.com/revrost/counterspell/internal/models"
)

const (
	defaultSessionSyncInterval = 5 * time.Second
	backendClaudeCode          = "claude-code"
	backendCodex               = "codex"
	sessionImportWindow        = 7 * 24 * time.Hour
)

var codexSetupMarkers = []string{
	"agents.md",
	"<environment_context>",
	"<collaboration_mode>",
	"<instructions>",
	"<permissions instructions>",
}

type importedMessage struct {
	Role       string
	Kind       string
	Content    string
	ToolName   string
	ToolCallID string
	RawJSON    string
	CreatedAt  int64
}

type SessionSyncer struct {
	repo *Repository

	pollInterval time.Duration
	stopCh       chan struct{}

	lastSeenMu sync.Mutex
	lastSeen   map[string]time.Time

	scanMu sync.Mutex
}

func NewSessionSyncer(repo *Repository) *SessionSyncer {
	return &SessionSyncer{
		repo:         repo,
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

		minAt, maxAt := sessionTimeBounds(messages)
		if maxAt == 0 {
			maxAt = info.ModTime().UnixMilli()
		}
		if minAt == 0 {
			minAt = maxAt
		}

		if !withinImportWindow(maxAt) {
			s.markFileSeen(path, info.ModTime())
			continue
		}

		if err := s.syncSession(ctx, backend, sessionID, messages, minAt, maxAt); err != nil {
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

func (s *SessionSyncer) syncSession(ctx context.Context, backend, sessionID string, messages []importedMessage, createdAt, lastMessageAt int64) error {
	session, err := s.repo.GetSessionByBackendExternal(ctx, backend, sessionID)
	if err != nil {
		return err
	}

	if session == nil {
		now := time.Now().UnixMilli()
		title := sessionTitle(messages)
		if title == "" {
			title = "Imported session"
		}
		newSession := &models.Session{
			ID:               shortuuid.New(),
			AgentBackend:     backend,
			ExternalID:       &sessionID,
			BackendSessionID: &sessionID,
			Title:            &title,
			MessageCount:     0,
			LastMessageAt:    &lastMessageAt,
			CreatedAt:        createdAt,
			UpdatedAt:        now,
		}
		session, err = s.repo.CreateSession(ctx, newSession)
		if err != nil {
			return err
		}
	} else {
		needsUpdate := false
		updateTitle := ""
		if session.Title == nil || strings.TrimSpace(*session.Title) == "" {
			updateTitle = sessionTitle(messages)
			if updateTitle != "" {
				needsUpdate = true
			}
		}
		updateBackendSessionID := ""
		if session.BackendSessionID == nil || *session.BackendSessionID == "" {
			updateBackendSessionID = sessionID
			needsUpdate = true
		}

		updatedLast := lastMessageAt
		if session.LastMessageAt != nil && *session.LastMessageAt > updatedLast {
			updatedLast = *session.LastMessageAt
		}

		if needsUpdate || session.LastMessageAt == nil || updatedLast != valueOrZero(session.LastMessageAt) {
			if updateTitle == "" && session.Title != nil {
				updateTitle = *session.Title
			}
			if updateBackendSessionID == "" && session.BackendSessionID != nil {
				updateBackendSessionID = *session.BackendSessionID
			}
			if err := s.repo.UpdateSession(ctx, session.ID, updateBackendSessionID, updateTitle, &updatedLast); err != nil {
				return err
			}
		}
	}

	start := int(session.MessageCount)
	if start < 0 {
		start = 0
	}
	if start >= len(messages) {
		return nil
	}

	for i, msg := range messages[start:] {
		sequence := int64(start + i)
		if msg.Kind == "" {
			msg.Kind = "text"
		}
		if strings.TrimSpace(msg.RawJSON) == "" {
			raw, _ := json.Marshal(map[string]any{
				"role":    msg.Role,
				"kind":    msg.Kind,
				"content": msg.Content,
			})
			msg.RawJSON = string(raw)
		}
		created := msg.CreatedAt
		if created == 0 {
			created = lastMessageAt
		}
		if err := s.repo.CreateSessionMessage(
			ctx,
			session.ID,
			sequence,
			msg.Role,
			msg.Kind,
			msg.Content,
			msg.ToolName,
			msg.ToolCallID,
			msg.RawJSON,
			created,
		); err != nil {
			return err
		}
	}

	return nil
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
		timestamp := extractTimestamp(event)

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
			if content := strings.TrimSpace(extractTextFromContent(event["content"])); content != "" {
				messages = append(messages, importedMessage{
					Role:      "system",
					Kind:      "system",
					Content:   content,
					RawJSON:   line,
					CreatedAt: timestamp,
				})
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
				messages = append(messages, importedMessage{
					Role:      role,
					Kind:      "text",
					Content:   content,
					RawJSON:   line,
					CreatedAt: timestamp,
				})
			}
		case "user_text":
			content := strings.TrimSpace(extractTextFromContent(event["content"]))
			if content != "" {
				messages = append(messages, importedMessage{
					Role:      "user",
					Kind:      "text",
					Content:   content,
					RawJSON:   line,
					CreatedAt: timestamp,
				})
			}
		case "tool_use":
			name, _ := event["name"].(string)
			id, _ := event["id"].(string)
			inputJSON := ""
			if input, ok := event["input"]; ok {
				if raw, err := json.Marshal(input); err == nil {
					inputJSON = string(raw)
				}
			}
			messages = append(messages, importedMessage{
				Role:       "assistant",
				Kind:       "tool_use",
				Content:    inputJSON,
				ToolName:   name,
				ToolCallID: id,
				RawJSON:    line,
				CreatedAt:  timestamp,
			})
		case "tool_result":
			content := strings.TrimSpace(extractTextFromContent(event["content"]))
			toolUseID, _ := event["tool_use_id"].(string)
			messages = append(messages, importedMessage{
				Role:       "tool",
				Kind:       "tool_result",
				Content:    content,
				ToolCallID: toolUseID,
				RawJSON:    line,
				CreatedAt:  timestamp,
			})
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
			markCodexSetupMessage(&msg)
			msg.RawJSON = line
			if msg.CreatedAt == 0 {
				msg.CreatedAt = extractTimestamp(payload)
			}
			messages = append(messages, msg)
			continue
		}
		if msg, ok := extractMessage(payload); ok {
			markCodexSetupMessage(&msg)
			msg.RawJSON = line
			if msg.CreatedAt == 0 {
				msg.CreatedAt = extractTimestamp(payload)
			}
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
				markCodexSetupMessage(&msg)
				rawJSON, _ := json.Marshal(msgMap)
				msg.RawJSON = string(rawJSON)
				if msg.CreatedAt == 0 {
					msg.CreatedAt = extractTimestamp(msgMap)
				}
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
	switch itemType {
	case "message":
		role, _ := payload["role"].(string)
		if role == "" {
			role = "assistant"
		}

		content := extractTextFromContent(payload["content"])
		content = strings.TrimSpace(content)
		if content == "" {
			// Some messages may only contain tool calls
			if toolCalls, ok := payload["tool_calls"].([]any); ok && len(toolCalls) > 0 {
				if toolMsg := extractToolCallMessage(toolCalls); toolMsg.Kind != "" {
					toolMsg.CreatedAt = extractTimestamp(payload)
					return toolMsg, true
				}
			}
			return importedMessage{}, false
		}

		return importedMessage{
			Role:      role,
			Kind:      "text",
			Content:   content,
			CreatedAt: extractTimestamp(payload),
		}, true
	case "tool_call", "function_call":
		name, _ := payload["name"].(string)
		id, _ := payload["id"].(string)
		args := ""
		if argVal, ok := payload["arguments"]; ok {
			if raw, err := json.Marshal(argVal); err == nil {
				args = string(raw)
			}
		}
		return importedMessage{
			Role:       "assistant",
			Kind:       "tool_use",
			Content:    args,
			ToolName:   name,
			ToolCallID: id,
			CreatedAt:  extractTimestamp(payload),
		}, true
	case "tool_result", "function_call_output":
		content := extractTextFromContent(payload["content"])
		if content == "" {
			if output, ok := payload["output"].(string); ok {
				content = output
			}
		}
		toolUseID, _ := payload["tool_call_id"].(string)
		return importedMessage{
			Role:       "tool",
			Kind:       "tool_result",
			Content:    strings.TrimSpace(content),
			ToolCallID: toolUseID,
			CreatedAt:  extractTimestamp(payload),
		}, true
	}

	return importedMessage{}, false
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
		if toolCalls, ok := payload["tool_calls"].([]any); ok && len(toolCalls) > 0 {
			toolMsg := extractToolCallMessage(toolCalls)
			if toolMsg.Kind != "" {
				toolMsg.CreatedAt = extractTimestamp(payload)
				return toolMsg, true
			}
		}
		return importedMessage{}, false
	}

	kind := "text"
	if role == "tool" {
		kind = "tool_result"
	}
	return importedMessage{
		Role:      role,
		Kind:      kind,
		Content:   content,
		CreatedAt: extractTimestamp(payload),
	}, true
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

func markCodexSetupMessage(msg *importedMessage) {
	if msg == nil {
		return
	}
	if strings.ToLower(msg.Role) != "user" {
		return
	}
	if !isCodexSetupContent(msg.Content) {
		return
	}
	msg.Role = "system"
	msg.Kind = "setup"
}

func isCodexSetupContent(content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}
	lower := strings.ToLower(content)
	for _, marker := range codexSetupMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
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

func extractTimestamp(payload map[string]any) int64 {
	if payload == nil {
		return 0
	}

	keys := []string{"created_at", "createdAt", "timestamp", "ts", "time", "date"}
	for _, key := range keys {
		if val, ok := payload[key]; ok {
			if ts := parseTimestamp(val); ts != 0 {
				return ts
			}
		}
	}

	if msg, ok := payload["message"].(map[string]any); ok {
		if ts := extractTimestamp(msg); ts != 0 {
			return ts
		}
	}

	return 0
}

func parseTimestamp(val any) int64 {
	switch v := val.(type) {
	case int64:
		return normalizeEpoch(v)
	case int:
		return normalizeEpoch(int64(v))
	case float64:
		return normalizeEpoch(int64(v))
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return normalizeEpoch(i)
		}
		if f, err := v.Float64(); err == nil {
			return normalizeEpoch(int64(f))
		}
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return 0
		}
		if num, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return normalizeEpoch(num)
		}
		if t, err := time.Parse(time.RFC3339Nano, trimmed); err == nil {
			return t.UnixMilli()
		}
		if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

func normalizeEpoch(val int64) int64 {
	if val <= 0 {
		return 0
	}
	if val > 1_000_000_000_000 {
		return val
	}
	if val > 1_000_000_000 {
		return val * 1000
	}
	return 0
}

func extractToolCallMessage(toolCalls []any) importedMessage {
	if len(toolCalls) == 0 {
		return importedMessage{}
	}
	callMap, ok := toolCalls[0].(map[string]any)
	if !ok {
		return importedMessage{}
	}

	toolName := ""
	toolCallID := ""
	args := ""

	if id, ok := callMap["id"].(string); ok {
		toolCallID = id
	}
	if name, ok := callMap["name"].(string); ok {
		toolName = name
	}
	if fn, ok := callMap["function"].(map[string]any); ok {
		if name, ok := fn["name"].(string); ok && name != "" {
			toolName = name
		}
		if arguments, ok := fn["arguments"]; ok {
			if raw, err := json.Marshal(arguments); err == nil {
				args = string(raw)
			}
		}
	}
	if args == "" {
		if arguments, ok := callMap["arguments"]; ok {
			if raw, err := json.Marshal(arguments); err == nil {
				args = string(raw)
			}
		}
	}

	if toolName == "" && args == "" && toolCallID == "" {
		return importedMessage{}
	}

	return importedMessage{
		Role:       "assistant",
		Kind:       "tool_use",
		Content:    args,
		ToolName:   toolName,
		ToolCallID: toolCallID,
	}
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

func sessionTitle(messages []importedMessage) string {
	for _, msg := range messages {
		if msg.Role == "user" && strings.TrimSpace(msg.Content) != "" {
			return truncateTitle(msg.Content)
		}
	}
	for _, msg := range messages {
		if strings.TrimSpace(msg.Content) != "" {
			return truncateTitle(msg.Content)
		}
	}
	return "Imported session"
}

func sessionTimeBounds(messages []importedMessage) (int64, int64) {
	var minAt int64
	var maxAt int64
	for _, msg := range messages {
		if msg.CreatedAt == 0 {
			continue
		}
		if minAt == 0 || msg.CreatedAt < minAt {
			minAt = msg.CreatedAt
		}
		if maxAt == 0 || msg.CreatedAt > maxAt {
			maxAt = msg.CreatedAt
		}
	}
	return minAt, maxAt
}

func withinImportWindow(lastMessageAt int64) bool {
	if lastMessageAt == 0 {
		return false
	}
	cutoff := time.Now().Add(-sessionImportWindow).UnixMilli()
	return lastMessageAt >= cutoff
}

func truncateTitle(title string) string {
	const maxLen = 120
	trimmed := strings.TrimSpace(title)
	if len(trimmed) <= maxLen {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:maxLen]) + "..."
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
