package clearcast

// RuntimeEvent is an event that is emitted by the runtime
// It could be either an error, tool execution result, agent transfer or final output
type RuntimeEvent interface {
	isEvent()
}

type Plan struct {
	Tool   string         `json:"tool"`
	Params map[string]any `json:"params"`
}

type PlanResultEvent struct {
	Type string `json:"type"`
	Plan []Plan `json:"plan"`
}

func (e *PlanResultEvent) isEvent() {}
func PlanResult(plan []Plan) RuntimeEvent {
	return &PlanResultEvent{
		Type: "plan_result",
		Plan: plan,
	}
}

type ErrorEvent struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

func (e *ErrorEvent) isEvent() {}

func Error(msg string) RuntimeEvent {
	return &ErrorEvent{
		Type:  "error",
		Error: msg,
	}
}

type ToolResultEvent struct {
	Type  string `json:"type"`
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func (e *ToolResultEvent) isEvent() {}
func ToolResult(data any, err error) RuntimeEvent {
	return &ToolResultEvent{
		Type:  "tool_result",
		Data:  data,
		Error: err.Error(),
	}
}

type TransferAgentEvent struct {
	Type    string `json:"type"`
	AgentID string `json:"agent_id"`
}

func (e *TransferAgentEvent) isEvent() {}
func TransferAgent(agentID string) RuntimeEvent {
	return &TransferAgentEvent{
		Type:    "transfer_agent",
		AgentID: agentID,
	}
}

type FinalEvent struct {
	Type   string `json:"type"`
	Output string `json:"output"`
}

func (e *FinalEvent) isEvent() {}
func Final(output string) RuntimeEvent {
	return &FinalEvent{
		Type:   "final",
		Output: output,
	}
}
