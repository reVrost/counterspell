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
	Run(ctx context.Context) (string, error)
	RunStream(ctx context.Context) <-chan RuntimeEvent
}

// manages the execution of agents or tools
type runtime struct {
	agents    Agents
	workspace map[string]any
	sess      *Session
}

type Tools map[string]*Tool
type Agents map[string]*Agent

func (t Tools) toMap() map[string]any {
	m := make(map[string]any)
	toolsList := make([]map[string]string, 0, len(t))
	for _, tool := range t {
		toolsList = append(toolsList, map[string]string{
			"id":          tool.ID,
			"description": tool.Description,
			"usage":       tool.Usage,
		})
	}
	m["tools"] = toolsList
	return m
}

// func (a Agents) StepStream(ctx context.Context, agentID string, sess *Session, opts ...StepOption) (<-chan ChatCompletionChunk, error) {
// 	agent, ok := a[agentID]
// 	if !ok {
// 		return nil, fmt.Errorf("agent not found: %s", agentID)
// 	}
// 	return agent.StepStream(ctx, sess.ToMap(), opts...)
// }

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

func WithSession(sess *Session) RuntimeOption {
	return func(r *runtime) {
		r.sess = sess
		r.workspace = sess.Workspace
	}
}

func NewRuntime(opts ...RuntimeOption) *runtime {
	rt := &runtime{
		agents:    make(map[string]*Agent),
		workspace: make(map[string]any),
	}
	for _, opt := range opts {
		opt(rt)
	}
	return rt
}

const maxTurns = 2

// runLoop is a plan-orchestrate execution style agent execution. It will
// iteratively plan and execute until a final answer is reached or max turns are exceeded.
func (r *runtime) runLoop(ctx context.Context, eventsChan chan RuntimeEvent, sess *Session) {
	rootAgent, ok := r.agents[sess.RootAgentID]
	if !ok {
		slog.Error("Agent not found", "agent", sess.RootAgentID)
		eventsChan <- Error(fmt.Sprintf("Agent not found: %s", sess.RootAgentID))
		return
	}

	r.workspace["next_hints"] = ""
	// The main execution loop. It will run for a maximum of `maxTurns`.
	for i := range maxTurns {
		r.workspace["iteration"] = i
		r.workspace["max_iterations"] = maxTurns
		// 1. PLAN: The agent thinks about what to do next based on the current workspace.
		planRes, err := rootAgent.Step(ctx, r.workspace, WithResponseFormat(ResponseFormat{
			Type: ResponseFormatTypeJSON,
		}))
		if err != nil {
			slog.Debug("Agent plan error", "agent", sess.RootAgentID, "error", err, "turn", i)
			eventsChan <- Error(fmt.Sprintf("Agent plan error: %s", err))
			return
		}

		plansEvent, err := DecodeMessage[*PlanResultEvent](planRes)
		if err != nil {
			slog.Debug("Agent decode plan error", "agent", sess.RootAgentID, "error", err, "content", planRes)
			eventsChan <- Error(fmt.Sprintf("Agent decode plan error: %s", err))
			return
		}

		if plansEvent != nil {
			eventsChan <- plansEvent
		}

		// 2. EXECUTE: The runtime executes the steps in the plan.
		for _, plan := range plansEvent.Plans {
			switch plan.Tool {
			case "plan": // This is the special tool indicating the end of a thought process.
				type Termination struct {
					Done        bool   `json:"done"`
					NextHints   string `json:"next_hints"` // Renamed from FinalOutput for clarity in non-done case
					FinalOutput string `json:"final_output"`
				}
				var termination Termination

				// More robust way to convert map[string]any to a struct
				jsonBytes, err := json.Marshal(plan.Params)
				if err != nil {
					eventsChan <- Error(fmt.Sprintf("Could not marshal plan termination: %s", err))
					return
				}
				if err := json.Unmarshal(jsonBytes, &termination); err != nil {
					eventsChan <- Error(fmt.Sprintf("Could not unmarshal plan termination: %s", err))
					return
				}

				if termination.Done {
					// The agent has finished. Send the final answer and exit.
					slog.Info("Agent has finished.", "agent", sess.RootAgentID)
					eventsChan <- Final(termination.FinalOutput)
					return // Success! Exit the function.
				}

				// Not done yet. We need to loop again.
				// Add the "next_hints" to the workspace so the next Plan() call has this context.
				slog.Debug("Continuing ReAct loop", "next_hints", termination.NextHints)
				r.workspace["next_hints"] = termination.NextHints

			default: // This is a standard tool call.
				slog.Debug("Executing tool", "agent", sess.RootAgentID, "tool", plan.Tool, "params", plan.Params)
				result, err := rootAgent.ExecuteTool(ctx, plan.Tool, plan.Params)
				if err != nil {
					slog.Debug("Tool execution error", "tool", plan.Tool, "error", err)
					// It's often better to feed the error back to the agent rather than halting.
					// This allows the agent to self-correct.
					r.workspace[plan.Tool] = fmt.Sprintf("error executing tool: %v", err)
					eventsChan <- ToolError(plan.Tool, plan.Params, err)
				} else {
					// Append successful result to the workspace for the next planning step.
					r.workspace[plan.Tool] = result
					eventsChan <- ToolResult(plan.Tool, result)
				}
			}
		}
	}

	// 3. HANDLE LOOP EXIT: If the for loop finishes, it means we hit maxTurns.
	slog.Warn("Agent exceeded max turns", "agent", sess.RootAgentID, "maxTurns", maxTurns)
	eventsChan <- Error(fmt.Sprintf("Agent %s exceeded maximum number of turns (%d)", sess.RootAgentID, maxTurns))
}

func (r *runtime) RunStream(ctx context.Context) <-chan RuntimeEvent {
	slog.Debug("Starting runtime stream", "agent", r.sess.RootAgentID)
	eventsChan := make(chan RuntimeEvent)

	go func() {
		defer close(eventsChan)
		/// TODO: record telemetry session
		r.runLoop(ctx, eventsChan, r.sess)
	}()
	return eventsChan
}

func (r *runtime) Run(ctx context.Context) (string, error) {
	eventsChan := r.RunStream(ctx)
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
