package orion

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"charm.land/fantasy"
)

// sessionAgent implements Agent interface with core
// agentic orchestration capabilities.
//
// NOTE: This is a simplified implementation. For full production use,
// you'll need to integrate with actual Fantasy library APIs
// for streaming, tool execution, and model communication.
type sessionAgent struct {
	mu               sync.RWMutex
	largeModel        fantasy.LanguageModel
	smallModel        fantasy.LanguageModel
	systemPrompt      string
	systemPromptPrefix string
	isSubAgent        bool
	disableAutoSummarize bool

	sessions    SessionService
	messages    MessageService
	tools       []fantasy.AgentTool
	eventBroker EventBroker[any]

	// Queue management
	messageQueue   map[string][]AgentCall
	activeRequests map[string]context.CancelFunc
}

// NewAgent creates a new agent with given options.
func NewAgent(opts AgentOptions) Agent {
	if opts.SystemPromptPrefix != "" {
		opts.SystemPrompt = opts.SystemPromptPrefix + "\n\n" + opts.SystemPrompt
	}

	agent := &sessionAgent{
		largeModel:        opts.LargeModel,
		smallModel:        opts.SmallModel,
		systemPrompt:      opts.SystemPrompt,
		systemPromptPrefix: opts.SystemPromptPrefix,
		isSubAgent:        opts.IsSubAgent,
		disableAutoSummarize: opts.DisableAutoSummarize,
		sessions:          opts.Sessions,
		messages:          opts.Messages,
		tools:             opts.Tools,
		eventBroker:       opts.EventBroker,
		messageQueue:      make(map[string][]AgentCall),
		activeRequests:     make(map[string]context.CancelFunc),
	}

	return agent
}

// Run executes an agent call with streaming support.
func (a *sessionAgent) Run(ctx context.Context, call AgentCall) (*fantasy.AgentResult, error) {
	// Validate
	if call.Prompt == "" {
		return nil, ErrEmptyPrompt
	}
	if call.SessionID == "" {
		return nil, ErrSessionMissing
	}

	// Queue message if busy
	if a.IsSessionBusy(call.SessionID) {
		a.mu.Lock()
		a.messageQueue[call.SessionID] = append(a.messageQueue[call.SessionID], call)
		a.mu.Unlock()
		return nil, ErrSessionBusy
	}

	// Mark session as busy
	a.mu.Lock()
	a.activeRequests[call.SessionID] = func() {}
	a.mu.Unlock()

	// Create cancelable context
	genCtx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.activeRequests[call.SessionID] = cancel
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		delete(a.activeRequests, call.SessionID)
		a.mu.Unlock()
	}()

	// Get or create session
	session, err := a.sessions.Get(ctx, call.SessionID)
	if err != nil {
		session, err = a.sessions.Create(ctx, call.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	// Create user message
	userMsg := NewMessage(call.SessionID, RoleUser)
	userMsg.Parts = append(userMsg.Parts, NewTextContent(call.Prompt))
	userMsg.AddFinish(FinishReasonEndTurn, "", "")

	_, err = a.messages.Create(ctx, call.SessionID, CreateMessageParams{
		Role:             userMsg.Role,
		Parts:            userMsg.Parts,
		IsSummaryMessage: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user message: %w", err)
	}

	// Add session to context
	ctx = context.WithValue(ctx, SessionIDContextKey, call.SessionID)

	// Create assistant message
	createdAssistantMsg, err := a.messages.Create(ctx, call.SessionID, CreateMessageParams{
		Role:             RoleAssistant,
		Parts:            []ContentPart{},
		IsSummaryMessage: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create assistant message: %w", err)
	}

	// Add context values
	ctx = context.WithValue(ctx, MessageIDContextKey, createdAssistantMsg.ID)

	// Create Fantasy agent
	agent := fantasy.NewAgent(
		a.largeModel,
		fantasy.WithSystemPrompt(a.systemPrompt),
		fantasy.WithTools(a.tools...),
	)

	// Execute agent stream
	// NOTE: This is where you integrate with Fantasy's actual streaming API
	// The callback implementations will depend on actual Fantasy types
	result, err := agent.Stream(genCtx, fantasy.AgentStreamCall{
		Prompt:           call.Prompt,
		Messages:         []fantasy.Message{}, // Build from history
		ProviderOptions:  call.ProviderOptions,
		MaxOutputTokens:  &call.MaxOutputTokens,
		TopP:             call.TopP,
		Temperature:      call.Temperature,
		PresencePenalty:  call.PresencePenalty,
		TopK:             call.TopK,
		FrequencyPenalty: call.FrequencyPenalty,
	})

	if err != nil {
		// Handle error
		if errors.Is(err, context.Canceled) {
			createdAssistantMsg.AddFinish(FinishReasonCanceled, "User canceled request", "")
		} else {
			createdAssistantMsg.AddFinish(FinishReasonError, "Error", err.Error())
		}
		_ = a.messages.Update(ctx, createdAssistantMsg)
		return nil, err
	}

	// Update assistant message with content
	// In full implementation, this would be streamed via callbacks
	createdAssistantMsg.AppendContent("[Agent response - implement streaming]")
	createdAssistantMsg.AddFinish(FinishReasonEndTurn, "", "")
	_ = a.messages.Update(ctx, createdAssistantMsg)

	// Update session usage
	if result.TotalUsage.InputTokens > 0 || result.TotalUsage.OutputTokens > 0 {
		session.PromptTokens = result.TotalUsage.InputTokens
		session.CompletionTokens = result.TotalUsage.OutputTokens
		// Simple cost calculation
		session.Cost += float64(result.TotalUsage.InputTokens)*0.000003 +
			float64(result.TotalUsage.OutputTokens)*0.000015
		_, err := a.sessions.Save(ctx, session)
		if err != nil {
			slog.Error("Failed to update session usage", "error", err)
		}
	}

	// Process queued messages
	a.mu.Lock()
	queued := a.messageQueue[call.SessionID]
	delete(a.messageQueue, call.SessionID)
	a.mu.Unlock()

	if len(queued) > 0 {
		// Process next queued message
		firstQueued := queued[0]
		remaining := queued[1:]
		a.mu.Lock()
		a.messageQueue[call.SessionID] = remaining
		a.mu.Unlock()

		go func() {
			_, err := a.Run(ctx, firstQueued)
			if err != nil {
				slog.Error("Failed to process queued message", "error", err)
			}
		}()
	}

	return result, nil
}

// SetModels updates large and small models used by agent.
func (a *sessionAgent) SetModels(large, small fantasy.LanguageModel) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.largeModel = large
	a.smallModel = small
}

// SetTools registers available tools with agent.
func (a *sessionAgent) SetTools(tools []fantasy.AgentTool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.tools = tools
}

// Cancel cancels an active request for given session ID.
func (a *sessionAgent) Cancel(sessionID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if cancel, ok := a.activeRequests[sessionID]; ok {
		cancel()
	}
	delete(a.activeRequests, sessionID)
}

// CancelAll cancels all active requests across all sessions.
func (a *sessionAgent) CancelAll() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for sessionID, cancel := range a.activeRequests {
		cancel()
		delete(a.activeRequests, sessionID)
	}
}

// IsSessionBusy checks if a session is currently processing a request.
func (a *sessionAgent) IsSessionBusy(sessionID string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	_, ok := a.activeRequests[sessionID]
	return ok
}

// IsBusy checks if any session is currently processing a request.
func (a *sessionAgent) IsBusy() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.activeRequests) > 0
}

// QueuedPrompts returns the number of queued prompts for a session.
func (a *sessionAgent) QueuedPrompts(sessionID string) int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.messageQueue[sessionID])
}

// QueuedPromptsList returns the list of queued prompts for a session.
func (a *sessionAgent) QueuedPromptsList(sessionID string) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	queued := a.messageQueue[sessionID]
	prompts := make([]string, len(queued))
	for i, call := range queued {
		prompts[i] = call.Prompt
	}
	return prompts
}

// ClearQueue clears all queued prompts for a session.
func (a *sessionAgent) ClearQueue(sessionID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.messageQueue, sessionID)
}

// Summarize creates a summary of the session conversation.
func (a *sessionAgent) Summarize(ctx context.Context, sessionID string, opts fantasy.ProviderOptions) error {
	if a.IsSessionBusy(sessionID) {
		return ErrSessionBusy
	}

	// This is where you implement summarization logic
	// For now, we'll mark it as a placeholder

	_, err := a.messages.Create(ctx, sessionID, CreateMessageParams{
		Role:             RoleAssistant,
		Parts:            []ContentPart{NewTextContent("[Summary placeholder - implement summarization]")},
		IsSummaryMessage: true,
	})

	return err
}

// Model returns the currently configured large model.
func (a *sessionAgent) Model() fantasy.LanguageModel {
	return a.largeModel
}
