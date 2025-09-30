package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/revrost/go-openrouter"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"

	"github.com/revrost/counterspell/pkg/clearcast"
)

// AgentDef defines the structure for an agent in the YAML configuration.
type AgentDef struct {
	ID                   string    `yaml:"id"`
	Mode                 string    `yaml:"mode"`
	Model                string    `yaml:"model"`
	Prompt               string    `yaml:"prompt"`
	AutoToolInstructions *bool     `yaml:"auto_tool_instructions,omitempty"` // nil = default (true), false = disable
	Tools                []ToolDef `yaml:"tools,omitempty"`                  // Tools specific to this agent
}

type ToolDef struct {
	ID string `yaml:"id"`
}

// SessionDef defines the structure for the session in the YAML configuration.
type SessionDef struct {
	RootAgentID string         `yaml:"root_agent_id"`
	Mission     string         `yaml:"mission"`
	Extra       map[string]any `yaml:",inline"` // Captures all other fields like topic, etc.
}

// OrchestrationFile defines the top-level structure of the YAML file.
type OrchestrationFile struct {
	Agents  []AgentDef `yaml:"agents"`
	Session SessionDef `yaml:"session"`
}

// EnableDebug sets slog to debug level with a default text handler
func EnableDebug() {
	// Create a new handler with Debug level enabled
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	// Replace the default logger with the debug-enabled one
	slog.SetDefault(slog.New(handler))
}

func runCommand(ctx context.Context, cmd *cli.Command) error {
	yamlFile := cmd.Args().First()
	if yamlFile == "" {
		return errors.New("usage: cspell run <yaml_file>")
	}

	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("failed to read file '%s': %w", yamlFile, err)
	}

	var config OrchestrationFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	b, _ := json.MarshalIndent(config, "", "\t")
	fmt.Printf("config :\n %s\n", string(b))

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY environment variable not set")
	}
	orClient := openrouter.NewClient(apiKey)
	llmProvider := clearcast.NewOpenRouterProvider(orClient)

	agents := make([]*clearcast.Agent, 0, len(config.Agents))
	for _, agentDef := range config.Agents {
		opts := []clearcast.AgentOption{}
		if agentDef.AutoToolInstructions != nil {
			opts = append(opts, clearcast.WithAutoToolInstructions(*agentDef.AutoToolInstructions))
		}

		// Add tools to agent
		if len(agentDef.Tools) > 0 {
			toolIDs := make([]string, 0, len(agentDef.Tools))
			for _, toolDef := range agentDef.Tools {
				toolIDs = append(toolIDs, toolDef.ID)
			}
			opts = append(opts, clearcast.WithDefaultTools(toolIDs...))
		}

		agent := clearcast.NewAgent(agentDef.ID, agentDef.Mode, agentDef.Model, agentDef.Prompt, llmProvider, opts...)
		agents = append(agents, agent)
	}

	rt := clearcast.NewRuntime(
		clearcast.WithAgents(agents...),
	)

	// Initialize workspace with session extra fields (like topic)
	workspace := make(map[string]any)
	for k, v := range config.Session.Extra {
		workspace[k] = v
	}

	sess := &clearcast.Session{
		RootAgentID:   config.Session.RootAgentID,
		Mission:       config.Session.Mission,
		Workspace:     workspace,
		Memory:        make(map[string]any),
		Messages:      []clearcast.Message{},
		CreatedAt:     time.Now(),
		MaxIterations: 10, // Default max iterations
	}

	eventsChan := rt.RunStream(ctx, sess)

	for event := range eventsChan {
		switch e := event.(type) {
		case *clearcast.PlanResultEvent:
			fmt.Println("=== Plan Created ===")
			for i, plan := range e.Plans {
				fmt.Printf("Step %d: [%s] %s with params %v\n", i+1, plan.Kind, plan.ID, plan.Params)
			}
		case *clearcast.AgentChoiceEvent:
			fmt.Println("\n=== Agent Output ===")
			fmt.Println(e.Content)
		case *clearcast.ErrorEvent:
			return fmt.Errorf("runtime error: %s", e.Error)
		case *clearcast.FinalEvent:
			fmt.Println("\n=== Execution Finished ===")
			fmt.Println(e.Output)
		default:
			// Optionally log unhandled event types
			// log.Printf("Received unhandled event type: %T\n", e)
		}
	}
	return nil
}

// main is the entry point for the CLI application.
// It takes a YAML file as input, initializes and runs a clearcast agent.
func main() {
	EnableDebug()
	cmd := &cli.Command{
		Name:  "cspell",
		Usage: "A CLI for running agents",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "run an agent from a yaml file",
				Action: runCommand,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
