// Package agent implements a simple coding agent loop.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lithammer/shortuuid/v4"
	"github.com/revrost/counterspell/internal/agent/tools"
	"github.com/revrost/counterspell/internal/llm"
)

// Message represents a single message in the conversation.
// The agent appends messages with both text and tool calls.
type Message struct {
	Role    string         `json:"role"` // "user" or "assistant"
	Content []ContentBlock `json:"content"`
}

// ContentBlock can be one of four types:
// - text: Assistant response text
// - thinking: Assistant hidden reasoning
// - tool_use: Assistant requesting to call a tool
// - tool_result: Result sent back after tool execution (from us to API)
type ContentBlock struct {
	// For all blocks
	Type string `json:"type"`

	// For text/thinking blocks
	Text string `json:"text,omitempty"`

	// For tool_use blocks (assistant calling a tool)
	Name  string         `json:"name,omitempty"`
	Input map[string]any `json:"input,omitempty"`
	ID    string         `json:"id,omitempty"` // Links to tool_result

	// For tool_result blocks (result sent back to assistant)
	ToolUseID string `json:"tool_use_id,omitempty"` // Links to tool_use ID
	Content   string `json:"content,omitempty"`     // Tool output
}

// RunnerOption customizes Runner behavior.
type RunnerOption func(*Runner)

// WithRunnerSystemPrompt overrides the default system prompt.
func WithRunnerSystemPrompt(prompt string) RunnerOption {
	return func(r *Runner) {
		if prompt != "" {
			r.systemPrompt = prompt
		}
	}
}

// Runner executes agent tasks with streaming output.
type Runner struct {
	provider       llm.Provider
	llmCaller      LLMCaller
	workDir        string
	systemPrompt   string
	finalMessage   string
	messageHistory []Message
	todoState      *tools.TodoState
	toolRegistry   *tools.Registry
	toolCtx        *tools.Context
}

// NewRunner creates a new agent runner.
func NewRunner(provider llm.Provider, workDir string, opts ...RunnerOption) *Runner {
	r := &Runner{
		provider:     provider,
		llmCaller:    NewLLMCaller(provider),
		workDir:      workDir,
		systemPrompt: fmt.Sprintf("You are a coding assistant. Work directory: %s. Be concise. Make changes directly.", workDir),
		todoState:    tools.NewTodoState(),
	}

	for _, opt := range opts {
		opt(r)
	}

	// Create tool registry with context
	toolCtx := &tools.Context{
		WorkDir:   workDir,
		TodoState: r.todoState,
	}
	r.toolCtx = toolCtx
	r.toolRegistry = tools.NewRegistry(toolCtx)

	return r
}

// GetFinalMessage returns the accumulated final message from the agent.
func (r *Runner) GetFinalMessage() string {
	return r.finalMessage
}

// GetMessageHistory returns the message history as JSON string.
func (r *Runner) GetMessageHistory() string {
	data, err := json.Marshal(r.messageHistory)
	if err != nil {
		return ""
	}
	return string(data)
}

// SetMessageHistory sets the initial message history from JSON string.
func (r *Runner) SetMessageHistory(historyJSON string) error {
	if historyJSON == "" {
		return nil
	}
	return json.Unmarshal([]byte(historyJSON), &r.messageHistory)
}

// GetTodos returns the current todo list as JSON string.
func (r *Runner) GetTodos() string {
	data, _ := json.Marshal(r.todoState.GetTodos())
	return string(data)
}

// GetTodoState returns the todo state.
func (r *Runner) GetTodoState() *tools.TodoState {
	return r.todoState
}

// Run executes the agent loop for a given task.
func (r *Runner) Run(ctx context.Context, task string) error {
	stream := r.Stream(ctx, task)
	return drainStream(ctx, stream)
}

// Stream executes the agent loop for a given task and returns a stream of events.
func (r *Runner) Stream(ctx context.Context, task string) *Stream {
	events := make(chan StreamEvent, 32)
	done := make(chan error, 1)
	todoEvents := make(chan []tools.TodoItem, 1)

	r.toolCtx.TodoEvents = todoEvents

	go func() {
		defer close(events)
		defer close(done)
		defer func() { r.toolCtx.TodoEvents = nil }()

		err := r.runWithMessage(ctx, task, false, events, todoEvents)
		done <- err
	}()

	return &Stream{Events: events, Done: done}
}

func drainStream(ctx context.Context, stream *Stream) error {
	if stream == nil {
		return nil
	}
	for stream.Events != nil || stream.Done != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-stream.Events:
			if !ok {
				stream.Events = nil
			}
		case err, ok := <-stream.Done:
			if !ok {
				stream.Done = nil
				continue
			}
			return err
		}
	}
	return nil
}

type blockBuilder struct {
	block ContentBlock
	args  strings.Builder
}

type messageBuilder struct {
	messageID string
	role      string
	blocks    []ContentBlock
	current   *blockBuilder
	toolCalls []ContentBlock
}

func (b *messageBuilder) ensureBlock(blockType string, block *ContentBlock) {
	if b.current != nil && b.current.block.Type == blockType {
		return
	}
	b.finalizeCurrent()
	bb := &blockBuilder{block: ContentBlock{Type: blockType}}
	if block != nil {
		bb.block = *block
		if bb.block.Type == "" {
			bb.block.Type = blockType
		}
	}
	b.current = bb
}

func (b *messageBuilder) appendDelta(blockType, delta string) {
	b.ensureBlock(blockType, nil)
	switch blockType {
	case "text", "thinking":
		b.current.block.Text += delta
	case "tool_use":
		b.current.args.WriteString(delta)
	}
}

func (b *messageBuilder) finalizeCurrent() *ContentBlock {
	if b.current == nil {
		return nil
	}
	block := b.current.block
	if block.Type == "tool_use" {
		argsJSON := strings.TrimSpace(b.current.args.String())
		if argsJSON != "" {
			input := map[string]any{}
			if err := json.Unmarshal([]byte(argsJSON), &input); err == nil {
				block.Input = input
			} else {
				block.Input = map[string]any{"raw": argsJSON}
			}
		}
	}
	b.blocks = append(b.blocks, block)
	if block.Type == "tool_use" {
		b.toolCalls = append(b.toolCalls, block)
	}
	b.current = nil
	return &block
}

func (b *messageBuilder) finalizeAll() {
	_ = b.finalizeCurrent()
}

// runWithMessage is the core loop that handles both new runs and continuations.
func (r *Runner) runWithMessage(ctx context.Context, userMessage string, isContinuation bool, events chan<- StreamEvent, todoEvents chan []tools.TodoItem) error {
	allTools := r.toolRegistry.All()

	// Use existing message history or start fresh
	messages := r.messageHistory
	if messages == nil {
		messages = []Message{}
	}

	if !isContinuation {
		r.finalMessage = ""
	}

	if userMessage != "" {
		// Add user message
		msg := Message{
			Role: "user",
			Content: []ContentBlock{{
				Type: "text",
				Text: userMessage,
			}},
		}
		messages = append(messages, msg)
	} else if len(messages) == 0 {
		return fmt.Errorf("agent: cannot start task with empty message")
	}

	todoDone := make(chan struct{})
	go func() {
		defer close(todoDone)
		for todos := range todoEvents {
			emitEvent(ctx, events, StreamEvent{Type: EventTodo, Todos: todos})
		}
	}()
	defer func() { <-todoDone }()
	defer close(todoEvents)

	// Agent loop
	for {
		select {
		case <-ctx.Done():
			r.messageHistory = messages
			return ctx.Err()
		default:
		}

		slog.Info("[RUNNER] Calling LLM API", "message_count", len(messages), "tool_count", len(allTools), "system_prompt", r.systemPrompt)
		stream, err := r.llmCaller.Stream(ctx, messages, allTools, r.systemPrompt)
		if err != nil {
			r.messageHistory = messages
			emitEvent(ctx, events, StreamEvent{Type: EventError, Error: err.Error()})
			slog.Error("[RUNNER] LLM API call failed", "error", err)
			return err
		}

		messageID := shortuuid.New()
		emitEvent(ctx, events, StreamEvent{Type: EventMessageStart, MessageID: messageID, Role: "assistant"})
		builder := &messageBuilder{messageID: messageID, role: "assistant"}
		messageEnded := false

		for stream.Events != nil || stream.Done != nil {
			select {
			case <-ctx.Done():
				stream.Events = nil
				stream.Done = nil
			case ev, ok := <-stream.Events:
				if !ok {
					stream.Events = nil
					continue
				}
				switch ev.Type {
				case LLMContentStart:
					builder.ensureBlock(ev.BlockType, ev.Block)
					emitEvent(ctx, events, StreamEvent{
						Type:      EventContentStart,
						MessageID: messageID,
						BlockType: ev.BlockType,
						Block:     ev.Block,
					})
				case LLMContentDelta:
					builder.appendDelta(ev.BlockType, ev.Delta)
					emitEvent(ctx, events, StreamEvent{
						Type:      EventContentDelta,
						MessageID: messageID,
						BlockType: ev.BlockType,
						Delta:     ev.Delta,
					})
				case LLMContentEnd:
					builder.ensureBlock(ev.BlockType, ev.Block)
					finished := builder.finalizeCurrent()
					emitEvent(ctx, events, StreamEvent{
						Type:      EventContentEnd,
						MessageID: messageID,
						BlockType: ev.BlockType,
						Block:     finished,
					})
				case LLMMessageEnd:
					messageEnded = true
				}
			case err, ok := <-stream.Done:
				if !ok {
					stream.Done = nil
					continue
				}
				if err != nil {
					emitEvent(ctx, events, StreamEvent{Type: EventError, Error: err.Error()})
					return err
				}
				stream.Done = nil
			}
		}

		builder.finalizeAll()
		assistantMsg := Message{Role: "assistant", Content: builder.blocks}
		messages = append(messages, assistantMsg)

		for _, block := range builder.blocks {
			if block.Type == "text" && block.Text != "" {
				r.finalMessage += block.Text
			}
		}

		if !messageEnded {
			messageEnded = true
		}
		if messageEnded {
			emitEvent(ctx, events, StreamEvent{Type: EventMessageEnd, MessageID: messageID, Role: "assistant"})
		}

		if len(builder.toolCalls) == 0 {
			slog.Info("[RUNNER] No more tools to run, completing task")
			break
		}

		toolResults := []ContentBlock{}
		for _, block := range builder.toolCalls {
			result := r.runTool(block.Name, block.Input, allTools)
			toolResults = append(toolResults, ContentBlock{
				Type:      "tool_result",
				ToolUseID: block.ID,
				Content:   result,
			})

			toolMsgID := shortuuid.New()
			emitEvent(ctx, events, StreamEvent{Type: EventMessageStart, MessageID: toolMsgID, Role: "user"})
			emitEvent(ctx, events, StreamEvent{
				Type:      EventContentStart,
				MessageID: toolMsgID,
				BlockType: "tool_result",
				Block: &ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   result,
				},
			})
			emitEvent(ctx, events, StreamEvent{
				Type:      EventContentEnd,
				MessageID: toolMsgID,
				BlockType: "tool_result",
				Block: &ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   result,
				},
			})
			emitEvent(ctx, events, StreamEvent{Type: EventMessageEnd, MessageID: toolMsgID, Role: "user"})
		}

		slog.Info("[RUNNER] Running %d tool result(s) through agent loop", "len_tool_results", len(toolResults))
		toolResultMsg := Message{Role: "user", Content: toolResults}
		messages = append(messages, toolResultMsg)
	}

	// Store message history for future continuations
	r.messageHistory = messages

	emitEvent(ctx, events, StreamEvent{Type: EventDone})
	return nil
}

func emitEvent(ctx context.Context, events chan<- StreamEvent, event StreamEvent) {
	select {
	case <-ctx.Done():
		return
	case events <- event:
	}
}

func (r *Runner) runTool(name string, args map[string]any, allTools map[string]tools.Tool) string {
	tool, ok := allTools[name]
	if !ok {
		return fmt.Sprintf("error: unknown tool %s", name)
	}
	defer func() {
		if rec := recover(); rec != nil {
			fmt.Printf("error in %s: %v\n", name, rec)
		}
	}()
	return tool.Func(args)
}
