package clearcast

import (
	"context"
	"fmt"
	"log/slog"
)

type EventKind string

// Runtime is an execution environment that can run tasks (agents or plain functions)
type Runtime interface {
	Run(ctx context.Context, prompt string) (string, error)
	RunStream(ctx context.Context, prompt string) <-chan RuntimeEvent
}

// manages the execution of agents or tools
type runtime struct {
	currentAgent  string
	maxIterations int
	iterations    int
	agents        map[string]Agent
}

func NewRuntime() *runtime {
	return &runtime{}
}

func (r *runtime) RunStream(ctx context.Context, prompt string) <-chan RuntimeEvent {
	slog.Debug("Starting runtime stream", "agent", r.currentAgent)
	eventsChan := make(chan RuntimeEvent)

	go func() {
		defer close(eventsChan)

		/// TODO: record telemetry session
		for {
			if r.iterations >= r.maxIterations {
				slog.Debug("Maximum iterations reached", "agent", r.currentAgent)
				eventsChan <- Final(fmt.Sprintf("Maximums iterations reached, %d", r.maxIterations))
				return
			}
			r.iterations++
			// immediate exit if ctrl c
			if err := ctx.Err(); err != nil {
				slog.Debug("Runtime stream cancelled", "agent", r.currentAgent, "error", err)
				eventsChan <- Final(fmt.Sprintf("Runtime stream cancelled, %s", err))
				return
			}
		}

	}()
	return nil
}

func (r *runtime) Run(ctx context.Context, prompt string) (string, error) {
	eventsChan := r.RunStream(ctx, prompt)

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
