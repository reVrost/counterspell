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
	agents        map[string]*Agent
	tools         map[string]*Tool
	workspace     map[string]any
}

func NewRuntime() *runtime {
	return &runtime{
		workspace: make(map[string]any),
	}
}

// runLoop is re-act style agent exxecution, it will recursively call itself untill the agent
// deemed the task is complete or if max iterations is reached
func (r *runtime) runLoop(ctx context.Context, eventsChan chan RuntimeEvent, sess *Session) {
	// 	for {
	// 		if r.iterations >= r.maxIterations {
	// 			slog.Debug("Maximum iterations reached", "agent", sess.RootAgentID)
	// 			eventsChan <- Final(fmt.Sprintf("Maximums iterations reached, %d", r.maxIterations))
	// 			return
	// 		}
	// 		r.iterations++
	// 		// immediate exit if context cancelled e.g ctrl c
	// 		if err := ctx.Err(); err != nil {
	// 			slog.Debug("Runtime stream cancelled", "agent", sess.RootAgentID, "error", err)
	// 			eventsChan <- Final(fmt.Sprintf("Runtime stream cancelled, %s", err))
	// 			return
	// 		}
	// 		// Sequential execution, start with current agent usually root
	// }
}

func DecodeMessage[T any](msg any) (T, error) {
	var data T
	err := json.Unmarshal([]byte(msg.(string)), &data)
	return data, err
}

// runPlan is a plan-orchestrate execution style agent execution, it will come up with a plan
// and execute it in a single step
func (r *runtime) runPlan(ctx context.Context, eventsChan chan RuntimeEvent, sess *Session) {
	planRes, err := r.agents[sess.RootAgentID].Step(ctx, sess.ToMap(), WithResponseFormat(ResponseFormat{
		Type: ResponseFormatTypeJSON,
	}))
	if err != nil {
		slog.Debug("Agent step error", "agent", sess.RootAgentID, "error", err)
		eventsChan <- Error(fmt.Sprintf("Agent step error: %s", err))
	}

	plansEvent, err := DecodeMessage[*PlanResultEvent](planRes)
	if err != nil {
		slog.Debug("Agent step error", "agent", sess.RootAgentID, "error", err)
		eventsChan <- Error(fmt.Sprintf("Agent step error: %s", err))
	}
	eventsChan <- plansEvent

	for _, plan := range plansEvent.Plans {
		// TODO: execute the plan
		slog.Debug("Executing plan", "agent", sess.RootAgentID, "plan", plan)
		switch plan.Kind {
		case PlanKindTool:
			result, err := r.tools[plan.ID].Execute(ctx, plan.Params)
			if err != nil {
				slog.Debug("Tool execution error", "tool", plan.ID, "error", err)
				eventsChan <- Error(fmt.Sprintf("Tool execution error: %s", err.Error()))
			}

			// append result to the workspace
			r.workspace[plan.ID] = result
		case PlanKindAgent:
			result, err := r.agents[plan.ID].Step(ctx, sess.ToMap())
			if err != nil {
				slog.Debug("Agent step error", "agent", plan.ID, "error", err)
				eventsChan <- Error(fmt.Sprintf("Agent step error: %s", err.Error()))
			}
			// stream result chan to events chan delta
			eventsChan <- &AgentChoiceEvent{
				Content: result.Content,
				Usage:   result.Usage,
			}
		default:
			slog.Debug("Unknown plan step", "agent", sess.RootAgentID, "step", plan)
			eventsChan <- Final(fmt.Sprintf("Unknown plan step %s", plan))
		}
	}
}

func (r *runtime) RunStream(ctx context.Context, sess *Session) <-chan RuntimeEvent {
	slog.Debug("Starting runtime stream", "agent", sess.RootAgentID)
	eventsChan := make(chan RuntimeEvent)

	go func() {
		defer close(eventsChan)
		/// TODO: record telemetry session
		agent := r.agents[sess.RootAgentID]
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
