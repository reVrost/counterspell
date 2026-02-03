package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/revrost/counterspell/internal/agent/tools"
)

// Compile-time interface check
var _ Backend = (*CodexBackend)(nil)

// ErrCodexBinaryPath is returned when the codex binary cannot be found.
var ErrCodexBinaryPath = errors.New("agent: codex binary not found in PATH")

// CodexBackend wraps the OpenAI Codex CLI as a Backend.
//
// It executes `codex exec --json` and normalizes the JSON event stream into
// StreamEvents for the UI.
type CodexBackend struct {
	binaryPath string
	workDir    string
	callback   StreamCallback
	apiKey     string
	baseURL    string
	model      string
	sessionID  string
	extraArgs  []string

	mu           sync.Mutex
	cmd          *exec.Cmd
	cancel       context.CancelFunc
	finalMessage string
	messages     []Message
}

// CodexOption configures a CodexBackend.
type CodexOption func(*CodexBackend)

// WithCodexBinaryPath sets the path to the codex binary.
// Defaults to "codex" (found via PATH).
func WithCodexBinaryPath(path string) CodexOption {
	return func(b *CodexBackend) {
		b.binaryPath = path
	}
}

// WithCodexWorkDir sets the working directory for the codex process.
func WithCodexWorkDir(dir string) CodexOption {
	return func(b *CodexBackend) {
		b.workDir = dir
	}
}

// WithCodexCallback sets the event streaming callback.
func WithCodexCallback(cb StreamCallback) CodexOption {
	return func(b *CodexBackend) {
		b.callback = cb
	}
}

// WithCodexAPIKey sets the API key for Codex CLI.
func WithCodexAPIKey(key string) CodexOption {
	return func(b *CodexBackend) {
		b.apiKey = key
	}
}

// WithCodexBaseURL sets a custom OpenAI-compatible API base URL.
func WithCodexBaseURL(url string) CodexOption {
	return func(b *CodexBackend) {
		b.baseURL = url
	}
}

// WithCodexModel sets the model to use.
func WithCodexModel(model string) CodexOption {
	return func(b *CodexBackend) {
		b.model = model
	}
}

// WithCodexSessionID sets the Codex thread/session ID to continue.
func WithCodexSessionID(sessionID string) CodexOption {
	return func(b *CodexBackend) {
		b.sessionID = sessionID
	}
}

// WithCodexExtraArgs appends extra CLI args to the codex command.
func WithCodexExtraArgs(args ...string) CodexOption {
	return func(b *CodexBackend) {
		b.extraArgs = append(b.extraArgs, args...)
	}
}

// NewCodexBackend creates a Codex CLI backend.
func NewCodexBackend(opts ...CodexOption) (*CodexBackend, error) {
	b := &CodexBackend{
		binaryPath: "codex",
		workDir:    ".",
	}
	for _, opt := range opts {
		opt(b)
	}

	// Verify binary exists
	if _, err := exec.LookPath(b.binaryPath); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCodexBinaryPath, b.binaryPath)
	}

	return b, nil
}

// --- Backend interface ---

// Run executes a new task via Codex CLI.
func (b *CodexBackend) Run(ctx context.Context, task string) error {
	return b.execute(ctx, task)
}

// Close terminates any running codex process.
func (b *CodexBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
	}
	if b.cmd != nil && b.cmd.Process != nil {
		return b.cmd.Process.Kill()
	}
	return nil
}

// --- Internal execution ---

func (b *CodexBackend) execute(ctx context.Context, prompt string) error {
	ctx, cancel := context.WithCancel(ctx)

	b.mu.Lock()
	b.cancel = cancel
	b.mu.Unlock()

	// Add user message to history and emit immediately
	b.mu.Lock()
	b.messages = append(b.messages, Message{
		Role:    "user",
		Content: []ContentBlock{{Type: "text", Text: prompt}},
	})
	b.mu.Unlock()
	b.emitMessages()

	cmd, err := b.buildCmd(ctx, prompt)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.cmd = cmd
	b.mu.Unlock()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	slog.Info("[CODEX] Starting command", "binary", b.binaryPath, "args", cmd.Args, "workdir", b.workDir, "api_key_len", len(b.apiKey), "base_url", b.baseURL)

	if err := cmd.Start(); err != nil {
		slog.Error("[CODEX] Failed to start command", "binary", b.binaryPath, "args", cmd.Args, "workdir", b.workDir, "error", err)
		return fmt.Errorf("start codex: %w", err)
	}

	// Parse output in background
	var wg sync.WaitGroup
	var stderrLines []string
	var stderrMu sync.Mutex

	wg.Go(func() {
		b.parseOutput(bufio.NewScanner(stdout))
	})

	wg.Go(func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				slog.Warn("[CODEX] stderr", "line", line)
				stderrMu.Lock()
				stderrLines = append(stderrLines, line)
				stderrMu.Unlock()
			}
		}
	})

	err = cmd.Wait()
	wg.Wait()

	if err != nil {
		stderrMu.Lock()
		stderrContent := strings.Join(stderrLines, "\n")
		stderrMu.Unlock()
		if stderrContent != "" {
			return fmt.Errorf("%w: %s", err, stderrContent)
		}
	}
	return err
}

func (b *CodexBackend) parseOutput(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Handle SSE-style output if present
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
		if line == "[DONE]" {
			b.emit(StreamEvent{Type: EventDone, Content: "Task completed"})
			continue
		}

		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Not JSON, emit as text
			b.emit(StreamEvent{Type: EventText, Content: line})
			continue
		}

		b.processCodexEvent(event)
	}
}

func (b *CodexBackend) processCodexEvent(event map[string]any) {
	eventType := getString(event, "type")

	switch eventType {
	case "thread.started":
		if threadID := getString(event, "thread_id"); threadID != "" {
			b.setSessionID(threadID)
		}
	case "turn.completed":
		b.emit(StreamEvent{Type: EventDone, Content: "Task completed"})
	case "turn.failed":
		b.emit(StreamEvent{Type: EventError, Content: extractCodexError(event)})
	case "error":
		b.emit(StreamEvent{Type: EventError, Content: extractCodexError(event)})
	case "item.started", "item.updated", "item.completed":
		item, _ := event["item"].(map[string]any)
		if item != nil {
			b.processCodexItem(eventType, item)
		}
	default:
		b.processCodexLegacyEvent(event)
	}
}

func (b *CodexBackend) processCodexItem(eventType string, item map[string]any) {
	itemType := getString(item, "type")
	if itemType == "" {
		itemType = getString(item, "item_type")
	}
	isCompleted := eventType == "item.completed"
	isUpdated := eventType == "item.updated"

	if isUpdated {
		switch itemType {
		case "assistant_message", "agent_message":
			text := extractTextFromContent(item["content"])
			if text == "" {
				text = getString(item, "text")
			}
			text = strings.TrimSpace(text)
			if text != "" {
				b.appendFinalText(text)
			}
		}
		return
	}

	switch itemType {
	case "assistant_message", "agent_message":
		if !isCompleted {
			return
		}
		text := extractTextFromContent(item["content"])
		if text == "" {
			text = getString(item, "text")
		}
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		b.appendMessage("assistant", []ContentBlock{{Type: "text", Text: text}})
		b.appendFinalText(text)
		return
	case "reasoning":
		// Intentionally ignored to avoid leaking reasoning content.
		return
	case "plan_update":
		// Treat plan updates as status text for now.
		if !isCompleted {
			return
		}
		text := extractTextFromContent(item["content"])
		if text == "" {
			text = getString(item, "text")
		}
		text = strings.TrimSpace(text)
		if text != "" {
			b.appendMessage("assistant", []ContentBlock{{Type: "text", Text: text}})
			b.appendFinalText(text)
		}
		return
	default:
		// Continue to tool handling below.
	}

	if !looksLikeCodexToolItem(itemType, item) {
		// Fallback: treat any text content as assistant output
		if isCompleted {
			text := extractTextFromContent(item["content"])
			if text == "" {
				text = getString(item, "text")
			}
			text = strings.TrimSpace(text)
			if text != "" {
				b.appendMessage("assistant", []ContentBlock{{Type: "text", Text: text}})
				b.appendFinalText(text)
			}
		}
		return
	}

	if isCompleted {
		b.emitCodexToolResult(itemType, item)
		return
	}

	b.emitCodexToolCall(itemType, item)
}

func (b *CodexBackend) processCodexLegacyEvent(event map[string]any) {
	eventType := getString(event, "type")

	switch eventType {
	case "session_meta":
		if payload, ok := event["payload"].(map[string]any); ok {
			if id := getString(payload, "id"); id != "" {
				b.setSessionID(id)
			}
		}
	case "response_item":
		if payload, ok := event["payload"].(map[string]any); ok {
			payloadType := getString(payload, "type")
			switch payloadType {
			case "message":
				role := getString(payload, "role")
				if role == "" {
					role = "assistant"
				}
				text := extractTextFromContent(payload["content"])
				text = strings.TrimSpace(text)
				if text != "" {
					b.appendMessage(role, []ContentBlock{{Type: "text", Text: text}})
					if role == "assistant" {
						b.appendFinalText(text)
					}
				}
			case "tool_call", "tool_use":
				b.emitCodexToolCall(payloadType, payload)
			case "tool_result", "tool_output":
				b.emitCodexToolResult(payloadType, payload)
			}
		}
	case "assistant":
		if message, ok := event["message"].(map[string]any); ok {
			text := extractTextFromContent(message["content"])
			text = strings.TrimSpace(text)
			if text != "" {
				b.appendMessage("assistant", []ContentBlock{{Type: "text", Text: text}})
				b.appendFinalText(text)
			}
		}
	case "assistant_message", "agent_message":
		text := extractTextFromContent(event["content"])
		if text == "" {
			text = getString(event, "text")
		}
		text = strings.TrimSpace(text)
		if text != "" {
			b.appendMessage("assistant", []ContentBlock{{Type: "text", Text: text}})
			b.appendFinalText(text)
		}
	case "assistant_message_delta", "agent_message_delta":
		delta := getString(event, "delta")
		if delta == "" {
			delta = getString(event, "text")
		}
		delta = strings.TrimSpace(delta)
		if delta != "" {
			b.appendFinalText(delta)
		}
	case "tool_call", "tool_use", "function_call":
		b.emitCodexToolCall(eventType, event)
	case "tool_result", "tool_output", "function_result":
		b.emitCodexToolResult(eventType, event)
	case "result":
		if isError, _ := event["is_error"].(bool); isError {
			b.emit(StreamEvent{Type: EventError, Content: extractCodexError(event)})
			return
		}
		b.emit(StreamEvent{Type: EventDone, Content: "Task completed"})
	default:
		// Streaming delta-style text events
		if delta := getString(event, "delta"); delta != "" {
			b.appendFinalText(delta)
			return
		}
		if text := getString(event, "text"); text != "" {
			b.appendMessage("assistant", []ContentBlock{{Type: "text", Text: text}})
			b.appendFinalText(text)
		}
	}
}

func (b *CodexBackend) emitCodexToolCall(itemType string, item map[string]any) {
	toolID := getString(item, "id")
	toolName, content, argsJSON := formatCodexToolCall(itemType, item)

	// Add tool use to messages
	b.mu.Lock()
	b.messages = append(b.messages, Message{
		Role: "assistant",
		Content: []ContentBlock{{
			Type:  "tool_use",
			Name:  toolName,
			ID:    toolID,
			Input: argsToMap(argsJSON, item),
		}},
	})
	b.mu.Unlock()
	b.emitMessages()

	b.emit(StreamEvent{
		Type:    EventTool,
		Tool:    toolName,
		Args:    argsJSON,
		Content: content,
	})
}

func (b *CodexBackend) emitCodexToolResult(itemType string, item map[string]any) {
	toolID := getString(item, "tool_call_id")
	if toolID == "" {
		toolID = getString(item, "tool_use_id")
	}
	if toolID == "" {
		toolID = getString(item, "id")
	}

	toolName, _, argsJSON := formatCodexToolCall(itemType, item)
	output := extractCodexToolResult(item)
	if output == "" {
		output = fmt.Sprintf("%s completed", toolName)
	}
	truncated := truncateDisplay(output, 200)

	// Add tool result to messages
	b.mu.Lock()
	b.messages = append(b.messages, Message{
		Role: "user",
		Content: []ContentBlock{{
			Type:      "tool_result",
			ToolUseID: toolID,
			Content:   output,
		}},
	})
	b.mu.Unlock()
	b.emitMessages()

	b.emit(StreamEvent{
		Type:    EventToolResult,
		Tool:    toolName,
		Args:    argsJSON,
		Content: truncated,
	})
}

func (b *CodexBackend) appendFinalText(text string) {
	if text == "" {
		return
	}
	b.mu.Lock()
	b.finalMessage += text
	b.mu.Unlock()
	b.emit(StreamEvent{Type: EventText, Content: text})
}

func (b *CodexBackend) appendMessage(role string, blocks []ContentBlock) {
	if len(blocks) == 0 {
		return
	}
	b.mu.Lock()
	b.messages = append(b.messages, Message{Role: role, Content: blocks})
	b.mu.Unlock()
	b.emitMessages()
}

func (b *CodexBackend) emit(event StreamEvent) {
	slog.Debug("[CODEX] emit", "type", event.Type, "content_len", len(event.Content), "has_callback", b.callback != nil)
	if b.callback != nil {
		b.callback(event)
	}
}

func (b *CodexBackend) emitMessages() {
	if b.callback == nil {
		slog.Debug("[CODEX] emitMessages skipped - no callback")
		return
	}
	b.mu.Lock()
	msgs := make([]Message, len(b.messages))
	copy(msgs, b.messages)
	b.mu.Unlock()

	data, err := json.Marshal(msgs)
	if err != nil {
		slog.Error("[CODEX] emitMessages failed to marshal", "error", err)
		return
	}
	slog.Debug("[CODEX] emitMessages", "msg_count", len(msgs), "data_len", len(data))
}

func (b *CodexBackend) setSessionID(sessionID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if sessionID == "" || b.sessionID == sessionID {
		return
	}
	b.sessionID = sessionID
	slog.Info("[CODEX] Session ID detected", "session_id", sessionID)
	b.emit(StreamEvent{Type: "session", SessionID: sessionID})
}

// --- Describable interface ---

// Info returns backend metadata.
func (b *CodexBackend) Info() BackendInfo {
	return BackendInfo{
		Type:    BackendCodex,
		Version: "1.0.0",
	}
}

// --- StatefulBackend interface ---

// GetState returns the conversation history as JSON.
func (b *CodexBackend) GetState() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	data, err := json.Marshal(b.messages)
	if err != nil {
		return ""
	}
	return string(data)
}

// RestoreState initializes the backend with previously saved state.
func (b *CodexBackend) RestoreState(stateJSON string) error {
	if stateJSON == "" {
		return nil
	}
	var msgs []Message
	if err := json.Unmarshal([]byte(stateJSON), &msgs); err != nil {
		return err
	}
	b.mu.Lock()
	b.messages = msgs
	b.mu.Unlock()
	b.emitMessages()
	return nil
}

// SessionID returns the current Codex session ID.
func (b *CodexBackend) SessionID() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sessionID
}

// --- IntrospectableBackend interface ---

// Messages returns the raw conversation history.
func (b *CodexBackend) Messages() []Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	result := make([]Message, len(b.messages))
	copy(result, b.messages)
	return result
}

// FinalMessage returns the accumulated assistant response text.
func (b *CodexBackend) FinalMessage() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.finalMessage
}

// Todos returns the current task list (empty for Codex backend).
func (b *CodexBackend) Todos() []tools.TodoItem {
	return nil
}

// buildCmd constructs the exec.Cmd for running Codex CLI.
func (b *CodexBackend) buildCmd(ctx context.Context, prompt string) (*exec.Cmd, error) {
	args := []string{"exec"}
	if b.sessionID != "" {
		args = append(args, "resume")
	}

	args = append(args, "--json", "--full-auto")
	if b.sessionID == "" {
		args = append(args, "--cd", b.workDir)
	}
	if b.model != "" {
		args = append(args, "--model", b.model)
	}
	if len(b.extraArgs) > 0 {
		args = append(args, b.extraArgs...)
	}
	if b.sessionID != "" {
		args = append(args, b.sessionID)
	}
	if prompt != "" {
		args = append(args, prompt)
	}

	cmd := exec.CommandContext(ctx, b.binaryPath, args...)
	cmd.Dir = b.workDir

	env := os.Environ()
	if b.apiKey != "" {
		env = append(env, "CODEX_API_KEY="+b.apiKey)
		env = append(env, "OPENAI_API_KEY="+b.apiKey)
	}
	if b.baseURL != "" {
		env = append(env, "OPENAI_BASE_URL="+b.baseURL)
		env = append(env, "OPENAI_API_BASE="+b.baseURL)
	}
	cmd.Env = env
	return cmd, nil
}

func looksLikeCodexToolItem(itemType string, item map[string]any) bool {
	switch itemType {
	case "command_execution", "file_change", "mcp_tool_call", "web_search", "tool_call", "tool_use", "tool_result", "tool_output":
		return true
	}
	if strings.Contains(itemType, "tool") || strings.Contains(itemType, "command") || strings.Contains(itemType, "file") || strings.Contains(itemType, "search") {
		return true
	}
	if _, ok := item["command"]; ok {
		return true
	}
	if _, ok := item["tool"]; ok {
		return true
	}
	if _, ok := item["tool_name"]; ok {
		return true
	}
	if _, ok := item["query"]; ok {
		return true
	}
	return false
}

func formatCodexToolCall(itemType string, item map[string]any) (toolName, content, argsJSON string) {
	toolName = itemType
	if name := getString(item, "name"); name != "" {
		toolName = name
	}
	if name := getString(item, "tool"); name != "" {
		toolName = name
	}
	if name := getString(item, "tool_name"); name != "" {
		toolName = name
	}

	if cmd := getString(item, "command"); cmd != "" {
		content = fmt.Sprintf("Running %s", cmd)
	} else if query := getString(item, "query"); query != "" {
		content = fmt.Sprintf("Searching: %s", query)
	} else {
		content = fmt.Sprintf("Running %s", toolName)
	}

	argsJSON = marshalJSON(item)
	return
}

func extractCodexToolResult(item map[string]any) string {
	if out := getString(item, "output"); out != "" {
		return out
	}
	if out := getString(item, "result"); out != "" {
		return out
	}
	if out := getString(item, "text"); out != "" {
		return out
	}
	if out := getString(item, "diff"); out != "" {
		return out
	}

	stdout := getString(item, "stdout")
	stderr := getString(item, "stderr")
	switch {
	case stdout != "" && stderr != "":
		return fmt.Sprintf("stdout:\n%s\n\nstderr:\n%s", stdout, stderr)
	case stdout != "":
		return stdout
	case stderr != "":
		return stderr
	}

	if files, ok := item["files"].([]any); ok && len(files) > 0 {
		return marshalJSON(files)
	}
	return ""
}

func extractCodexError(event map[string]any) string {
	if msg := getString(event, "message"); msg != "" {
		return msg
	}
	if msg := getString(event, "error"); msg != "" {
		return msg
	}
	if payload, ok := event["error"].(map[string]any); ok {
		if msg := getString(payload, "message"); msg != "" {
			return msg
		}
		return marshalJSON(payload)
	}
	return "Codex execution failed"
}

func getString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}

func marshalJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func truncateDisplay(content string, limit int) string {
	if limit <= 0 || len(content) <= limit {
		return content
	}
	return content[:limit] + "..."
}

func argsToMap(argsJSON string, fallback map[string]any) map[string]any {
	if argsJSON == "" {
		return fallback
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &parsed); err == nil {
		return parsed
	}
	return fallback
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
