package clearcast

import "context"

type ExecuteFunc func(ctx context.Context, params map[string]any) (any, error)

type Tool struct {
	ID          string      `json:"id"`
	Description string      `json:"description"`
	Execute     ExecuteFunc `json:"execute"`
	Usage       string      `json:"usage"`
}

// func (t *Tool) Execute(ctx context.Context, params map[string]any) (RuntimeEvent, error) {
// 	return nil, nil
// }
