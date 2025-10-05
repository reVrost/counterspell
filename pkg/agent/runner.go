package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/cagent/pkg/agent"
	"github.com/docker/cagent/pkg/config"
	"github.com/docker/cagent/pkg/model/provider"
	"github.com/docker/cagent/pkg/environment"
	"github.com/docker/cagent/pkg/model/provider/options"
	"github.com/docker/cagent/pkg/runtime"
	"github.com/docker/cagent/pkg/session"
	"github.com/docker/cagent/pkg/team"
	latest "github.com/docker/cagent/pkg/config/v2"
)

type Runner struct {
	cfgPath string

	openRouterCfg any
	telemetryCfg  any
	schedulerCfg  any
}

type RunnerOption func(*Runner)

func WithOpenRouter(v any) RunnerOption { return func(r *Runner) { r.openRouterCfg = v } }
func WithTelemetry(v any) RunnerOption  { return func(r *Runner) { r.telemetryCfg = v } }
func ScheduleWithAnts(v any) RunnerOption {
	return func(r *Runner) { r.schedulerCfg = v }
}

func NewRunner(configPath string, opts ...RunnerOption) *Runner {
	r := &Runner{cfgPath: configPath}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Runner) Run() error {
	cfg, err := config.LoadConfigSecure(r.cfgPath, "")
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	// Build agents from cfg.Agents
	agentMap := make(map[string]*agent.Agent)
	for name, a := range cfg.Agents {
		// Build provider(s) for this agent from its model field
		var provs []provider.Provider
		for _, modelName := range splitComma(a.Model) {
			m, ok := cfg.Models[modelName]
			if !ok {
				return fmt.Errorf("model %s not defined in models", modelName)
			}
			p, err := provider.New(context.Background(), &m, environment.NewOsEnvProvider(), options.WithGateway(""))
			if err != nil {
				return fmt.Errorf("creating provider for %s: %w", modelName, err)
			}
			provs = append(provs, p)
		}
		ag := agent.New(name, a.Instruction, agent.WithDescription(a.Description))
		for _, p := range provs {
			ag = agent.New(name, a.Instruction, agent.WithDescription(a.Description), agent.WithModel(p))
		}
		if a.AddDate {
			ag = agent.New(name, a.Instruction, agent.WithDescription(a.Description), agent.WithModel(provs[0]), agent.WithAddDate(true))
		}
		if a.AddEnvironmentInfo {
			ag = agent.New(name, a.Instruction, agent.WithDescription(a.Description), agent.WithModel(provs[0]), agent.WithAddEnvironmentInfo(true))
		}
		if a.MaxIterations > 0 {
			ag = agent.New(name, a.Instruction, agent.WithDescription(a.Description), agent.WithModel(provs[0]), agent.WithMaxIterations(a.MaxIterations))
		}
		agentMap[name] = ag
	}
	// Wire sub-agents
	for name, a := range cfg.Agents {
		if len(a.SubAgents) == 0 {
			continue
		}
		subs := make([]*agent.Agent, 0, len(a.SubAgents))
		for _, s := range a.SubAgents {
			if sub, ok := agentMap[s]; ok {
				subs = append(subs, sub)
			}
		}
		base := agentMap[name]
		if base == nil {
			continue
		}
		agentMap[name] = agent.New(name, base.Instruction(), agent.WithDescription(base.Description()), agent.WithModel(base.Model()), agent.WithSubAgents(subs...))
	}

	teamOpts := []team.Opt{team.WithID("team")}
	agents := make([]*agent.Agent, 0, len(agentMap))
	for _, ag := range agentMap {
		agents = append(agents, ag)
	}
	teamOpts = append(teamOpts, team.WithAgents(agents...))
	teamObj := team.New(teamOpts...)

	rt, err := runtime.New(teamObj)
	if err != nil {
		return fmt.Errorf("could not create runtime: %w", err)
	}

	s := session.New(
		session.WithSystemMessage(""),
		session.WithUserMessage("", "Follow the default instructions"),
	)
	_, err = rt.Run(context.Background(), s)
	if err != nil {
		return fmt.Errorf("error running agent: %w", err)
	}
	return nil
}

func splitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func agentMapModelFallback(cfg *latest.Config, modelStr string) provider.Provider {
	names := splitComma(modelStr)
	if len(names) == 0 {
		return nil
	}
	m, ok := cfg.Models[names[0]]
	if !ok {
		return nil
	}
	p, _ := provider.New(context.Background(), &m, environment.NewOsEnvProvider(), options.WithGateway(""))
	return p
}

func RunAgentWithCagent(configPath string) error {
	return NewRunner(configPath).Run()
}
