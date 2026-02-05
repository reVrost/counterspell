package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/lithammer/shortuuid/v4"
	"github.com/revrost/counterspell/internal/agent/codex"
	"github.com/revrost/counterspell/internal/agent/tools"
)

// Compile-time interface check
var _ Backend = (*CodexBackend)(nil)

// ErrCodexBinaryPath is returned when the codex binary cannot be found.
var ErrCodexBinaryPath = errors.New("agent: codex binary not found in PATH")

// CodexBackend wraps the OpenAI Codex CLI as a Backend.
//
// It uses the Go Codex SDK (codex.Codex) and normalizes the JSON event stream
// into StreamEvents for the UI.
type CodexBackend struct {
	binaryPath   string
	workDir      string
	apiKey       string
	baseURL      string
	model        string
	threadID     string
	systemPrompt string
	logStream    bool

	client *codex.Codex

	streamCtx           context.Context
	events              chan<- StreamEvent
	streamMsgID         string
	streamMsgRole       string
	streamText          string
	streamThinkingMsgID string
	streamThinkingRole  string
	streamThinkingText  string

	mu           sync.Mutex
	cancel       context.CancelFunc
	finalMessage string
	messages     []Message
	todos        []tools.TodoItem
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
		b.threadID = sessionID
	}
}

// WithCodexSystemPrompt sets an initial system prompt prefix.
func WithCodexSystemPrompt(prompt string) CodexOption {
	return func(b *CodexBackend) {
		b.systemPrompt = prompt
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

	if env := strings.TrimSpace(os.Getenv("CODEX_DEBUG_STREAM")); env != "" && env != "0" && env != "false" {
		b.logStream = true
	}

	b.client = codex.New(codex.Options{
		CodexPathOverride: b.binaryPath,
		BaseURL:           b.baseURL,
		APIKey:            b.apiKey,
	})

	return b, nil
}

// --- Backend interface ---

// Run executes a new task via Codex CLI.
func (b *CodexBackend) Run(ctx context.Context, task string) error {
	stream := b.Stream(ctx, task)
	return drainStream(ctx, stream)
}

// Stream executes a task via Codex CLI and returns a stream of events.
func (b *CodexBackend) Stream(ctx context.Context, task string) *Stream {
	events := make(chan StreamEvent, 32)
	done := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(done)
		b.setStream(ctx, events)
		defer b.clearStream()

		done <- b.execute(ctx, task)
	}()

	return &Stream{Events: events, Done: done}
}

// Close terminates any running codex process.
func (b *CodexBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
	}
	return nil
}

func (b *CodexBackend) setStream(ctx context.Context, events chan<- StreamEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.streamCtx = ctx
	b.events = events
}

func (b *CodexBackend) clearStream() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.streamCtx = nil
	b.events = nil
}

// --- Internal execution ---

func (b *CodexBackend) execute(ctx context.Context, prompt string) error {
	ctx, cancel := context.WithCancel(ctx)

	b.mu.Lock()
	b.cancel = cancel
	b.mu.Unlock()

	execPrompt := prompt
	if b.systemPrompt != "" && b.threadID == "" {
		execPrompt = b.systemPrompt + "\n\n" + prompt
	}

	b.mu.Lock()
	b.messages = append(b.messages, Message{
		Role:    "user",
		Content: []ContentBlock{{Type: "text", Text: prompt}},
	})
	b.mu.Unlock()

	threadOptions := codex.ThreadOptions{
		Model: b.model,
	}
	if b.threadID == "" {
		threadOptions.WorkingDirectory = b.workDir
	}

	thread := b.client.StartThread(threadOptions)
	if b.threadID != "" {
		thread = b.client.ResumeThread(b.threadID, threadOptions)
	}

	stream := thread.RunStreamed(ctx, execPrompt, codex.TurnOptions{})
	for event := range stream.Events {
		if b.logStream {
			slog.Info("[CODEX] event", "type", event.Type, "thread_id", event.ThreadID, "raw", event.Raw)
		}
		if event.Raw != nil {
			b.processCodexEvent(event.Raw)
			continue
		}
		b.handleThreadEventFallback(event)
	}
	return <-stream.Done
}

func (b *CodexBackend) handleThreadEventFallback(event codex.ThreadEvent) {
	switch event.Type {
	case "thread.started":
		if event.ThreadID != "" {
			b.setThreadID(event.ThreadID)
		}
	case "turn.failed":
		if event.Error != nil && event.Error.Message != "" {
			b.emit(StreamEvent{Type: EventError, Error: event.Error.Message})
			return
		}
		b.emit(StreamEvent{Type: EventError, Error: "Codex execution failed"})
	case "error":
		if event.Message != "" {
			b.emit(StreamEvent{Type: EventError, Error: event.Message})
			return
		}
		b.emit(StreamEvent{Type: EventError, Error: "Codex execution failed"})
	}
}

func (b *CodexBackend) processCodexEvent(event map[string]any) {
	eventType := getString(event, "type")

	switch eventType {
	case "thread.started":
		if threadID := getString(event, "thread_id"); threadID != "" {
			b.setThreadID(threadID)
		}
	case "turn.completed":
		b.finalizeStreamThinking("assistant")
		b.finalizeStreamText("assistant")
		b.emit(StreamEvent{Type: EventDone})
	case "turn.failed":
		b.finalizeStreamThinking("assistant")
		b.emit(StreamEvent{Type: EventError, Error: extractCodexError(event)})
	case "error":
		b.finalizeStreamThinking("assistant")
		b.emit(StreamEvent{Type: EventError, Error: extractCodexError(event)})
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
			if text == "" {
				text = getString(item, "delta")
			}
			text = strings.TrimSpace(text)
			if text != "" {
				b.appendStreamText(text)
			}
		case "reasoning":
			text := extractTextFromContent(item["content"])
			if text == "" {
				text = getString(item, "text")
			}
			if text == "" {
				text = getString(item, "delta")
			}
			text = strings.TrimSpace(text)
			if text != "" {
				b.appendStreamThinking(text)
			}
		}
		return
	}

	switch itemType {
	case "assistant_message", "agent_message":
		if !isCompleted {
			return
		}
		b.finalizeStreamThinking("assistant")
		text := extractTextFromContent(item["content"])
		if text == "" {
			text = getString(item, "text")
		}
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		b.finalizeStreamText("assistant")
		b.emitTextMessage("assistant", text)
		return
	case "reasoning":
		if !isCompleted {
			return
		}
		text := extractTextFromContent(item["content"])
		if text == "" {
			text = getString(item, "text")
		}
		text = strings.TrimSpace(text)
		if text != "" {
			b.appendStreamThinking(text)
		}
		b.finalizeStreamThinking("assistant")
		return
	case "plan_update":
		// Treat plan updates as status text for now.
		if !isCompleted {
			return
		}
		b.finalizeStreamThinking("assistant")
		text := extractTextFromContent(item["content"])
		if text == "" {
			text = getString(item, "text")
		}
		text = strings.TrimSpace(text)
		if text != "" {
			b.finalizeStreamText("assistant")
			b.emitTextMessage("assistant", text)
		}
		return
	case "todo_list":
		if !isCompleted && !isUpdated {
			return
		}
		todos := parseCodexTodoList(item)
		if len(todos) == 0 {
			return
		}
		b.setTodos(todos)
		b.emit(StreamEvent{Type: EventTodo, Todos: todos})
		return
	default:
		// Continue to tool handling below.
	}

	if !looksLikeCodexToolItem(itemType, item) {
		if isCompleted {
			b.finalizeStreamThinking("assistant")
			text := extractTextFromContent(item["content"])
			if text == "" {
				text = getString(item, "text")
			}
			text = strings.TrimSpace(text)
			if text != "" {
				b.finalizeStreamText("assistant")
				b.emitTextMessage("assistant", text)
			}
		}
		return
	}

	if isCompleted {
		b.finalizeStreamThinking("assistant")
		b.finalizeStreamText("assistant")
		b.emitCodexToolResult(itemType, item)
		return
	}

	b.finalizeStreamThinking("assistant")
	b.finalizeStreamText("assistant")
	b.emitCodexToolCall(itemType, item)
}

func (b *CodexBackend) processCodexLegacyEvent(event map[string]any) {
	eventType := getString(event, "type")

	switch eventType {
	case "session_meta":
		if payload, ok := event["payload"].(map[string]any); ok {
			if id := getString(payload, "id"); id != "" {
				b.setThreadID(id)
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
					b.finalizeStreamThinking("assistant")
					b.finalizeStreamText("assistant")
					b.emitTextMessage(role, text)
				}
			case "tool_call", "tool_use":
				b.finalizeStreamThinking("assistant")
				b.finalizeStreamText("assistant")
				b.emitCodexToolCall(payloadType, payload)
			case "tool_result", "tool_output":
				b.finalizeStreamThinking("assistant")
				b.finalizeStreamText("assistant")
				b.emitCodexToolResult(payloadType, payload)
			}
		}
	case "assistant":
		if message, ok := event["message"].(map[string]any); ok {
			text := extractTextFromContent(message["content"])
			text = strings.TrimSpace(text)
			if text != "" {
				b.finalizeStreamThinking("assistant")
				b.finalizeStreamText("assistant")
				b.emitTextMessage("assistant", text)
			}
		}
	case "reasoning":
		text := extractTextFromContent(event["content"])
		if text == "" {
			text = getString(event, "text")
		}
		text = strings.TrimSpace(text)
		if text != "" {
			b.appendStreamThinking(text)
			b.finalizeStreamThinking("assistant")
		}
	case "reasoning_delta":
		delta := getString(event, "delta")
		if delta == "" {
			delta = getString(event, "text")
		}
		delta = strings.TrimSpace(delta)
		if delta != "" {
			b.appendStreamThinking(delta)
		}
	case "assistant_message", "agent_message":
		text := extractTextFromContent(event["content"])
		if text == "" {
			text = getString(event, "text")
		}
		text = strings.TrimSpace(text)
		if text != "" {
			b.finalizeStreamText("assistant")
			b.emitTextMessage("assistant", text)
		}
	case "assistant_message_delta", "agent_message_delta":
		delta := getString(event, "delta")
		if delta == "" {
			delta = getString(event, "text")
		}
		delta = strings.TrimSpace(delta)
		if delta != "" {
			b.appendStreamText(delta)
		}
	case "tool_call", "tool_use", "function_call":
		b.finalizeStreamThinking("assistant")
		b.finalizeStreamText("assistant")
		b.emitCodexToolCall(eventType, event)
	case "tool_result", "tool_output", "function_result":
		b.finalizeStreamThinking("assistant")
		b.finalizeStreamText("assistant")
		b.emitCodexToolResult(eventType, event)
	case "result":
		if isError, _ := event["is_error"].(bool); isError {
			b.emit(StreamEvent{Type: EventError, Error: extractCodexError(event)})
			return
		}
		b.finalizeStreamThinking("assistant")
		b.finalizeStreamText("assistant")
		b.emit(StreamEvent{Type: EventDone})
	default:
		if delta := getString(event, "delta"); delta != "" {
			b.appendStreamText(delta)
			return
		}
		if text := getString(event, "text"); text != "" {
			b.finalizeStreamThinking("assistant")
			b.finalizeStreamText("assistant")
			b.emitTextMessage("assistant", text)
		}
	}
}

func (b *CodexBackend) emitCodexToolCall(itemType string, item map[string]any) {
	toolID := getString(item, "id")
	toolName, _, argsJSON := formatCodexToolCall(itemType, item)

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
	msgID := shortuuid.New()
	block := &ContentBlock{Type: "tool_use", Name: toolName, ID: toolID, Input: argsToMap(argsJSON, item)}
	b.emit(StreamEvent{Type: EventMessageStart, MessageID: msgID, Role: "assistant"})
	b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "tool_use", Block: block})
	b.emit(StreamEvent{Type: EventContentEnd, MessageID: msgID, BlockType: "tool_use", Block: block})
	b.emit(StreamEvent{Type: EventMessageEnd, MessageID: msgID, Role: "assistant"})
}

func (b *CodexBackend) emitCodexToolResult(itemType string, item map[string]any) {
	toolID := getString(item, "tool_call_id")
	if toolID == "" {
		toolID = getString(item, "tool_use_id")
	}
	if toolID == "" {
		toolID = getString(item, "id")
	}

	toolName, _, _ := formatCodexToolCall(itemType, item)
	output := extractCodexToolResult(item)
	if output == "" {
		output = fmt.Sprintf("%s completed", toolName)
	}
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
	msgID := shortuuid.New()
	block := &ContentBlock{Type: "tool_result", ToolUseID: toolID, Content: output}
	b.emit(StreamEvent{Type: EventMessageStart, MessageID: msgID, Role: "user"})
	b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "tool_result", Block: block})
	b.emit(StreamEvent{Type: EventContentEnd, MessageID: msgID, BlockType: "tool_result", Block: block})
	b.emit(StreamEvent{Type: EventMessageEnd, MessageID: msgID, Role: "user"})
}

func (b *CodexBackend) emit(event StreamEvent) {
	b.mu.Lock()
	ctx := b.streamCtx
	events := b.events
	b.mu.Unlock()
	if events == nil {
		return
	}
	if ctx == nil {
		events <- event
		return
	}
	select {
	case <-ctx.Done():
		return
	case events <- event:
	}
}

func (b *CodexBackend) startStreamMessage(role string) string {
	b.mu.Lock()
	if b.streamMsgID != "" {
		id := b.streamMsgID
		b.mu.Unlock()
		return id
	}
	id := shortuuid.New()
	b.streamMsgID = id
	b.streamMsgRole = role
	b.mu.Unlock()
	b.emit(StreamEvent{Type: EventMessageStart, MessageID: id, Role: role})
	return id
}

func (b *CodexBackend) endStreamMessage() {
	b.mu.Lock()
	id := b.streamMsgID
	role := b.streamMsgRole
	b.streamMsgID = ""
	b.streamMsgRole = ""
	b.mu.Unlock()
	if id != "" {
		b.emit(StreamEvent{Type: EventMessageEnd, MessageID: id, Role: role})
	}
}

func (b *CodexBackend) startStreamThinking(role string) string {
	b.mu.Lock()
	if b.streamThinkingMsgID != "" {
		id := b.streamThinkingMsgID
		b.mu.Unlock()
		return id
	}
	id := shortuuid.New()
	b.streamThinkingMsgID = id
	b.streamThinkingRole = role
	b.mu.Unlock()
	b.emit(StreamEvent{Type: EventMessageStart, MessageID: id, Role: role})
	return id
}

func (b *CodexBackend) endStreamThinking() {
	b.mu.Lock()
	id := b.streamThinkingMsgID
	role := b.streamThinkingRole
	b.streamThinkingMsgID = ""
	b.streamThinkingRole = ""
	b.mu.Unlock()
	if id != "" {
		b.emit(StreamEvent{Type: EventMessageEnd, MessageID: id, Role: role})
	}
}

func (b *CodexBackend) appendStreamText(delta string) {
	if delta == "" {
		return
	}
	msgID := b.startStreamMessage("assistant")
	start := false
	b.mu.Lock()
	if b.streamText == "" {
		start = true
	}
	b.streamText += delta
	b.finalMessage += delta
	b.mu.Unlock()
	if start {
		b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "text", Block: &ContentBlock{Type: "text"}})
	}
	b.emit(StreamEvent{Type: EventContentDelta, MessageID: msgID, BlockType: "text", Delta: delta})
}

func (b *CodexBackend) appendStreamThinking(delta string) {
	if delta == "" {
		return
	}
	msgID := b.startStreamThinking("assistant")
	start := false
	b.mu.Lock()
	if b.streamThinkingText == "" {
		start = true
	}
	b.streamThinkingText += delta
	b.mu.Unlock()
	if start {
		b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "thinking", Block: &ContentBlock{Type: "thinking"}})
	}
	b.emit(StreamEvent{Type: EventContentDelta, MessageID: msgID, BlockType: "thinking", Delta: delta})
}

func (b *CodexBackend) finalizeStreamText(role string) {
	b.mu.Lock()
	text := b.streamText
	b.streamText = ""
	msgID := b.streamMsgID
	b.mu.Unlock()
	if text == "" {
		return
	}
	b.appendMessage(role, []ContentBlock{{Type: "text", Text: text}})
	b.emit(StreamEvent{Type: EventContentEnd, MessageID: msgID, BlockType: "text", Block: &ContentBlock{Type: "text", Text: text}})
	b.endStreamMessage()
}

func (b *CodexBackend) finalizeStreamThinking(role string) {
	b.mu.Lock()
	text := b.streamThinkingText
	b.streamThinkingText = ""
	msgID := b.streamThinkingMsgID
	b.mu.Unlock()
	if text == "" {
		return
	}
	b.appendMessage(role, []ContentBlock{{Type: "thinking", Text: text}})
	b.emit(StreamEvent{Type: EventContentEnd, MessageID: msgID, BlockType: "thinking", Block: &ContentBlock{Type: "thinking", Text: text}})
	b.endStreamThinking()
}

func (b *CodexBackend) appendMessage(role string, blocks []ContentBlock) {
	if len(blocks) == 0 {
		return
	}
	b.mu.Lock()
	b.messages = append(b.messages, Message{Role: role, Content: blocks})
	b.mu.Unlock()
}

func (b *CodexBackend) emitTextMessage(role, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	b.appendMessage(role, []ContentBlock{{Type: "text", Text: text}})
	if role == "assistant" {
		b.mu.Lock()
		b.finalMessage += text
		b.mu.Unlock()
	}
	msgID := shortuuid.New()
	b.emit(StreamEvent{Type: EventMessageStart, MessageID: msgID, Role: role})
	b.emit(StreamEvent{Type: EventContentStart, MessageID: msgID, BlockType: "text", Block: &ContentBlock{Type: "text"}})
	b.emit(StreamEvent{Type: EventContentEnd, MessageID: msgID, BlockType: "text", Block: &ContentBlock{Type: "text", Text: text}})
	b.emit(StreamEvent{Type: EventMessageEnd, MessageID: msgID, Role: role})
}

func (b *CodexBackend) setThreadID(threadID string) {
	b.mu.Lock()
	if threadID == "" || b.threadID == threadID {
		b.mu.Unlock()
		return
	}
	b.threadID = threadID
	b.mu.Unlock()
	slog.Info("[CODEX] Thread ID detected", "thread_id", threadID)
	b.emit(StreamEvent{Type: EventSession, SessionID: threadID})
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
	return nil
}

// SessionID returns the current Codex session ID.
func (b *CodexBackend) SessionID() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.threadID
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
	b.mu.Lock()
	defer b.mu.Unlock()
	result := make([]tools.TodoItem, len(b.todos))
	copy(result, b.todos)
	return result
}

func (b *CodexBackend) setTodos(items []tools.TodoItem) {
	b.mu.Lock()
	b.todos = make([]tools.TodoItem, len(items))
	copy(b.todos, items)
	b.mu.Unlock()
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
	if out := getString(item, "aggregated_output"); out != "" {
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

func parseCodexTodoList(item map[string]any) []tools.TodoItem {
	rawItems, ok := item["items"].([]any)
	if !ok {
		return nil
	}
	todos := make([]tools.TodoItem, 0, len(rawItems))
	for _, raw := range rawItems {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		text, _ := entry["text"].(string)
		if strings.TrimSpace(text) == "" {
			continue
		}
		completed, _ := entry["completed"].(bool)
		status := tools.TodoStatusPending
		if completed {
			status = tools.TodoStatusCompleted
		}
		todos = append(todos, tools.TodoItem{
			Content:    text,
			Status:     status,
			ActiveForm: "",
		})
	}
	return todos
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
	blocks, ok := content.([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(blocks))
	for _, block := range blocks {
		entry, ok := block.(map[string]any)
		if !ok {
			continue
		}
		text, _ := entry["text"].(string)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "")
}
