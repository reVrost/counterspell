package clearcast

import "context"

type Tool struct {
	ID string
}

func (t *Tool) Execute(ctx context.Context, params map[string]any) (RuntimeEvent, error) {
	return nil, nil
}
