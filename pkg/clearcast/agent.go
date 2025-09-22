package clearcast

import "context"

const AgentModePlan = "plan"
const AgentModeLoop = "loop"

type Agent struct {
	ID    string
	Model string
	Mode  string
}

func NewAgent() *Agent {
	return &Agent{}
}

func (a *Agent) Step(ctx context.Context, input string) (RuntimeEvent, error) {
	return nil, nil
}
