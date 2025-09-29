package clearcast

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

type EventKind string

// Runtime is an execution environment that can run tasks (agents or plain functions)
type Runtime interface {
	Run(ctx context.Context, sess *Session) (string, error)
	RunStream(ctx context.Context, sess *Session) <-chan RuntimeEvent
}

// manages the execution of agents or tools
type runtime struct {
	maxIterations int
	iterations    int
	agents        Agents
	tools         Tools
	workspace     map[string]any
}

type Tools map[string]*Tool
type Agents map[string]*Agent

func (a Agents) Step(ctx context.Context, agentID string, sess *Session, opts ...StepOption) (ChatCompletionResponse, error) {
	agent, ok := a[agentID]
	if !ok {
		return ChatCompletionResponse{}, fmt.Errorf("agent not found: %s", agentID)
	}
	return agent.Step(ctx, sess.ToMap(), opts...)
}

func (a Agents) StepStream(ctx context.Context, agentID string, sess *Session, opts ...StepOption) (<-chan ChatCompletionChunk, error) {
	agent, ok := a[agentID]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}
	return agent.StepStream(ctx, sess.ToMap(), opts...)
}

func (t Tools) Execute(ctx context.Context, toolID string, params map[string]any) (any, error) {
	tool, ok := t[toolID]
	if !ok {
		return ChatCompletionResponse{}, fmt.Errorf("tool not found: %s", toolID)
	}
	return tool.Execute(ctx, params)
}

type RuntimeOption func(*runtime)

func WithAgents(agents ...*Agent) RuntimeOption {
	return func(r *runtime) {
		for _, agent := range agents {
			r.agents[agent.ID] = agent
		}
		slog.Debug("Agents added", "agents", r.agents)
	}
}

func WithTools(tools ...*Tool) RuntimeOption {
	return func(r *runtime) {
		for _, tool := range tools {
			r.tools[tool.ID] = tool
		}
	}
}

func NewRuntime(opts ...RuntimeOption) *runtime {
	rt := &runtime{
		agents:    make(map[string]*Agent),
		tools:     make(map[string]*Tool),
		workspace: make(map[string]any),
	}
	for _, opt := range opts {
		opt(rt)
	}
	return rt
}

// runLoop is re-act style agent exxecution, it will recursively call itself untill the agent
// deemed the task is complete or if max iterations is reached
func (r *runtime) runLoop(ctx context.Context, eventsChan chan RuntimeEvent, sess *Session) {
	for {
		if r.iterations >= r.maxIterations {
			slog.Debug("Maximum iterations reached", "agent", sess.RootAgentID)
			eventsChan <- Final(fmt.Sprintf("Maximum iterations reached, %d", r.maxIterations))
			return
		}
		r.iterations++
		// immediate exit if context cancelled e.g ctrl c
		if err := ctx.Err(); err != nil {
			slog.Debug("Runtime stream cancelled", "agent", sess.RootAgentID, "error", err)
			eventsChan <- Final(fmt.Sprintf("Runtime stream cancelled, %s", err))
			return
		}
		// Agent takes a step
		result, err := r.agents.Step(ctx, sess.RootAgentID, sess)
		if err != nil {
			slog.Debug("Agent step error", "agent", sess.RootAgentID, "error", err)
			eventsChan <- Error(fmt.Sprintf("Agent step error: %s", err))
			return
		}

		eventsChan <- &AgentChoiceEvent{
			Content: result.Content,
			Usage:   result.Usage,
		}

		var action struct {
			Tool        string         `json:"tool"`
			Params      map[string]any `json:"params"`
			FinalAnswer string         `json:"final_answer"`
		}

		if err := json.Unmarshal([]byte(result.Content), &action); err != nil {
			slog.Debug("Could not decode agent action, assuming it's a final answer", "error", err, "content", result.Content)
			eventsChan <- Final(result.Content)
			return
		}

		if action.FinalAnswer != "" {
			eventsChan <- Final(action.FinalAnswer)
			return
		}

		if action.Tool != "" {
			tool, ok := r.tools[action.Tool]
			if !ok {
				errorMsg := fmt.Sprintf("Tool not found: %s", action.Tool)
				slog.Debug(errorMsg)
				sess.Messages = append(sess.Messages, Message{Role: "assistant", Content: result.Content})
				sess.Messages = append(sess.Messages, Message{Role: "tool", Content: fmt.Sprintf("Error: %s", errorMsg)})
				continue
			}

			toolResult, err := tool.Execute(ctx, action.Params)
			if err != nil {
				errorMsg := fmt.Sprintf("Tool execution error: %s", err)
				slog.Debug(errorMsg)
				sess.Messages = append(sess.Messages, Message{Role: "assistant", Content: result.Content})
				sess.Messages = append(sess.Messages, Message{Role: "tool", Content: fmt.Sprintf("Error: %s", errorMsg)})
				continue
			}

			sess.Messages = append(sess.Messages, Message{Role: "assistant", Content: result.Content})

			toolResultBytes, _ := json.Marshal(toolResult)
			observation := string(toolResultBytes)

			sess.Messages = append(sess.Messages, Message{Role: "tool", Content: observation})
		} else {
			// No tool, no final answer. What to do? Assume it's the final answer.
			eventsChan <- Final(result.Content)
			return
		}
	}
}

// runPlan is a plan-orchestrate execution style agent execution, it will come up with a plan
// and execute it in a single step
func (r *runtime) runPlan(ctx context.Context, eventsChan chan RuntimeEvent, sess *Session) {
	planRes, err := r.agents.Step(ctx, sess.RootAgentID, sess, WithResponseFormat(ResponseFormat{
		Type: ResponseFormatTypeJSON,
	}))
	if err != nil {
		slog.Debug("Agent step error", "agent", sess.RootAgentID, "error", err, "content", planRes)
		eventsChan <- Error(fmt.Sprintf("Agent step error: %s", err))
	}

	plansEvent, err := DecodeMessage[*PlanResultEvent](planRes)
	if err != nil {
		slog.Debug("Agent decode step error", "agent", sess.RootAgentID, "error", err, "content", planRes)
		eventsChan <- Error(fmt.Sprintf("Agent step error: %s", err))
	}
	eventsChan <- plansEvent

	for _, plan := range plansEvent.Plans {
		// TODO: execute the plan
		slog.Debug("Executing plan", "agent", sess.RootAgentID, "plan", plan)
		switch plan.Kind {
		case PlanKindTool:
			result, err := r.tools.Execute(ctx, plan.ID, plan.Params)
			if err != nil {
				slog.Debug("Tool execution error", "tool", plan.ID, "error", err)
				eventsChan <- Error(fmt.Sprintf("Tool execution error: %s", err.Error()))
			}

			// append result to the workspace
			r.workspace[plan.ID] = result
		case PlanKindAgent:
			result, err := r.agents.Step(ctx, plan.ID, sess)
			if err != nil {
				slog.Debug("Agent step error", "agent", plan.ID, "error", err)
				eventsChan <- Error(fmt.Sprintf("Agent step error: %s", err.Error()))
			}
			// stream result chan to events chan delta
			eventsChan <- &AgentChoiceEvent{
				Content: result.Content,
				Usage:   result.Usage,
			}
			r.workspace[plan.ID] = result.Content
		default:
			slog.Debug("Unknown plan step", "agent", sess.RootAgentID, "step", plan)
			eventsChan <- Final(fmt.Sprintf("Unknown plan step %s", plan))
		}
	}
	eventsChan <- Final(fmt.Sprintf("Agent %s finished", sess.RootAgentID))
}

func (r *runtime) RunStream(ctx context.Context, sess *Session) <-chan RuntimeEvent {
	slog.Debug("Starting runtime stream", "agent", sess.RootAgentID)
	eventsChan := make(chan RuntimeEvent)

	go func() {
		defer close(eventsChan)
		/// TODO: record telemetry session
		agent, ok := r.agents[sess.RootAgentID]
		if !ok {
			slog.Error("Agent not found", "agent", sess.RootAgentID)
			eventsChan <- Error(fmt.Sprintf("Agent not found: %s", sess.RootAgentID))
			return
		}
		if agent.mode == AgentModePlan {
			r.runPlan(ctx, eventsChan, sess)
		} else {
			r.runLoop(ctx, eventsChan, sess)
		}
	}()
	return eventsChan
}

func (r *runtime) Run(ctx context.Context, sess *Session) (string, error) {
	eventsChan := r.RunStream(ctx, sess)
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case event := <-eventsChan:
			if errEvent, ok := event.(*ErrorEvent); ok {
				return "", fmt.Errorf("%s", errEvent.Error)
			}
			if finalEvent, ok := event.(*FinalEvent); ok {
				return finalEvent.Output, nil
			}
		}
	}
}
