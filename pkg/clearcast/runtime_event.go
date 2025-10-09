package clearcast

// RuntimeEvent is an event that is emitted by the runtime
// It could be either an error, tool execution result, agent transfer or final output
type RuntimeEvent interface {
	isEvent()
}

type Plan struct {
	// Kind can be either "tool" or "agent"
	// Kind   string         `json:"kind"`
	Tool   string         `json:"tool"`
	Params map[string]any `json:"params"`
}

type PlanResultEvent struct {
	Plans []Plan `json:"plans"`
}

func (e *PlanResultEvent) isEvent() {}
func PlanResult(plan []Plan) RuntimeEvent {
	return &PlanResultEvent{
		Plans: plan,
	}
}

type AgentChoiceEvent struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
}

func (e *AgentChoiceEvent) isEvent() {}

type AgentReasoningEvent struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
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

type ToolErrorEvent struct {
	Tool   string `json:"tool"`
	Error  string `json:"error"`
	Params any    `json:"params"`
}

func (e *ToolErrorEvent) isEvent() {}
func ToolError(tool string, params any, err error) RuntimeEvent {
	return &ToolErrorEvent{
		Tool:   tool,
		Error:  err.Error(),
		Params: params,
	}
}

type ToolResultEvent struct {
	Tool   string `json:"tool"`
	Result any    `json:"result"`
	Error  string `json:"error"`
}

func (e *ToolResultEvent) isEvent() {}
func ToolResult(key string, result any) RuntimeEvent {
	return &ToolResultEvent{
		Result: result,
	}
}

type TransferAgentEvent struct {
	AgentID string `json:"agent_id"`
}

func (e *TransferAgentEvent) isEvent() {}
func TransferAgent(agentID string) RuntimeEvent {
	return &TransferAgentEvent{
		AgentID: agentID,
	}
}

type FinalEvent struct {
	Output string `json:"output"`
}

func (e *FinalEvent) isEvent() {}
func Final(output string) RuntimeEvent {
	return &FinalEvent{
		Output: output,
	}
}
