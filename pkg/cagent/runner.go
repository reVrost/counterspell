package cagent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/cagent/pkg/config"
	"github.com/docker/cagent/pkg/runtime"
	"github.com/docker/cagent/pkg/session"
	"github.com/docker/cagent/pkg/teamloader"
)

// configSource is internal struct holding configuration input.
type configSource struct {
	path string
	raw  []byte
}

// Runner is the orchestrator for loading configuration and running agents.
type Runner struct {
	src           configSource
	openRouterCfg any
	telemetryCfg  any
	schedulerCfg  any
}

var runConfig config.RuntimeConfig

// RunnerOption allows optional configuration of the Runner.
type RunnerOption func(*Runner)

func WithOpenRouter(v any) RunnerOption   { return func(r *Runner) { r.openRouterCfg = v } }
func WithTelemetry(v any) RunnerOption    { return func(r *Runner) { r.telemetryCfg = v } }
func ScheduleWithAnts(v any) RunnerOption { return func(r *Runner) { r.schedulerCfg = v } }

// --- Constructors ---

func NewRunnerFromPath(path string, opts ...RunnerOption) *Runner {
	r := &Runner{src: configSource{path: path}}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func NewRunnerFromString(cfg string, opts ...RunnerOption) *Runner {
	return NewRunnerFromBytes([]byte(cfg), opts...)
}

func NewRunnerFromBytes(cfg []byte, opts ...RunnerOption) *Runner {
	r := &Runner{src: configSource{raw: cfg}}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// --- Main Logic ---

func (r *Runner) Run() error {
	if r.src.path == "" {
		return errors.New("no configuration path provided")
	}
	abs, err := filepath.Abs(r.src.path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		return fmt.Errorf("config path not found: %s", abs)
	}

	ctx := context.Background()
	agents, err := teamloader.Load(ctx, abs, runConfig)
	if err != nil {
		return fmt.Errorf("failed to load team: %w", err)
	}
	defer agents.StopToolSets()

	rt, err := runtime.New(agents, runtime.WithCurrentAgent("root"))
	if err != nil {
		return fmt.Errorf("could not create runtime: %w", err)
	}

	sess := session.New()
	sess.AddMessage(session.UserMessage(abs, "Follow the default instructions"))

	st := rt.RunStream(ctx, sess)

	for event := range st {
		switch event := event.(type) {
		case *runtime.ErrorEvent:
			return fmt.Errorf("runtime error: %s", event.Error)
		default:
			fmt.Printf("%s\n", event)
		}
	}

	return nil
}
