package clearcast

import (
	"context"
	"fmt"
	"log/slog"
)

type EventKind string

// Runtime is an execution environment that can run tasks (agents or plain functions)
type Runtime interface {
	Run(ctx context.Context, rootAgentID, userInput string) (string, error)
	RunStream(ctx context.Context, rootAgentID, userInput string) <-chan RuntimeEvent
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
func (r *runtime) runLoop(ctx context.Context,
	eventsChan chan RuntimeEvent, rootAgentID, userInput string) {
	// 	for {
	// 		if r.iterations >= r.maxIterations {
	// 			slog.Debug("Maximum iterations reached", "agent", rootAgentID)
	// 			eventsChan <- Final(fmt.Sprintf("Maximums iterations reached, %d", r.maxIterations))
	// 			return
	// 		}
	// 		r.iterations++
	// 		// immediate exit if context cancelled e.g ctrl c
	// 		if err := ctx.Err(); err != nil {
	// 			slog.Debug("Runtime stream cancelled", "agent", rootAgentID, "error", err)
	// 			eventsChan <- Final(fmt.Sprintf("Runtime stream cancelled, %s", err))
	// 			return
	// 		}
	// 		// Sequential execution, start with current agent usually root
	// }
}

// runPlan is a plan-orchestrate execution style agent execution, it will come up with a plan
// and execute it in a single step
func (r *runtime) runPlan(ctx context.Context,
	eventsChan chan RuntimeEvent, rootAgentID, userInput string) {
	plan, err := r.agents[rootAgentID].Step(ctx, userInput)
	if err != nil {
		slog.Debug("Agent step error", "agent", rootAgentID, "error", err)
		eventsChan <- Error(fmt.Sprintf("Agent step error: %s", err))
	}

	planResult, ok := plan.(*PlanResultEvent)
	if !ok {
		// not a plan result event, we can't execute it
		slog.Debug("Agent step not a plan result event", "agent", rootAgentID, "plan", plan)
		eventsChan <- Error(fmt.Sprintf("Agent step not a plan result event: %s", plan))
	}

	for _, stepInPlan := range planResult.Plan {
		// TODO: execute the plan
		slog.Debug("Executing plan", "agent", rootAgentID, "plan", plan)
		switch stepInPlan.Kind {
		case PlanKindTool:
			result, err := r.tools[stepInPlan.ID].Execute(ctx, stepInPlan.Params)
			if err != nil {
				slog.Debug("Tool execution error", "tool", stepInPlan.ID, "error", err)
				eventsChan <- Error(fmt.Sprintf("Tool execution error: %s", err.Error()))
			}

			// append result to the workspace
			r.workspace[stepInPlan.ID] = result
		case PlanKindAgent:
			result, err := r.agents[stepInPlan.ID].Step(ctx, userInput)
			if err != nil {
				slog.Debug("Agent step error", "agent", stepInPlan.ID, "error", err)
				eventsChan <- Error(fmt.Sprintf("Agent step error: %s", err.Error()))
			}
			// stream result chan to events chan delta
			eventsChan <- result
		default:
			slog.Debug("Unknown plan step", "agent", rootAgentID, "step", stepInPlan)
			eventsChan <- Final(fmt.Sprintf("Unknown plan step %s", stepInPlan))
		}
	}
}

func (r *runtime) RunStream(ctx context.Context, rootAgentID, userInput string) <-chan RuntimeEvent {
	slog.Debug("Starting runtime stream", "agent", rootAgentID)
	eventsChan := make(chan RuntimeEvent)

	go func() {
		defer close(eventsChan)
		/// TODO: record telemetry session
		agent := r.agents[rootAgentID]
		if agent.mode == AgentModePlan {
			r.runPlan(ctx, eventsChan, rootAgentID, userInput)
		} else {
			r.runLoop(ctx, eventsChan, rootAgentID, userInput)
		}
	}()
	return eventsChan
}

func (r *runtime) Run(ctx context.Context, rootAgentID, userInput string) (string, error) {
	eventsChan := r.RunStream(ctx, rootAgentID, userInput)
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
