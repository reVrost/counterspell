package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/revrost/go-openrouter"
	"gopkg.in/yaml.v3"

	"github.com/revrost/counterspell/pkg/clearcast"
)

// AgentDef defines the structure for an agent in the YAML configuration.
type AgentDef struct {
	ID     string `yaml:"id"`
	Mode   string `yaml:"mode"`
	Model  string `yaml:"model"`
	Prompt string `yaml:"prompt"`
}

// SessionDef defines the structure for the session in the YAML configuration.
type SessionDef struct {
	RootAgentID string `yaml:"root_agent_id"`
	Mission     string `yaml:"mission"`
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

// main is the entry point for the CLI application.
// It takes a YAML file as input, initializes and runs a clearcast agent.
func main() {
	EnableDebug()
	flag.Parse()
	yamlFile := flag.Arg(0)
	if yamlFile == "" {
		log.Fatal("Usage: cspell run <yaml_file>")
	}

	data, err := os.ReadFile(yamlFile)
	if err != nil {
		log.Fatalf("Failed to read file '%s': %v", yamlFile, err)
	}

	var config OrchestrationFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	b, _ := json.MarshalIndent(config, "", "\t")
	fmt.Printf("config :\n %s\n", string(b))

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable not set")
	}
	orClient := openrouter.NewClient(apiKey)
	llmProvider := clearcast.NewOpenRouterProvider(orClient)

	agents := make([]*clearcast.Agent, 0, len(config.Agents))
	for _, agentDef := range config.Agents {
		agent := clearcast.NewAgent(agentDef.ID, agentDef.Mode, agentDef.Model, agentDef.Prompt, llmProvider)
		agents = append(agents, agent)
	}

	rt := clearcast.NewRuntime(
		clearcast.WithAgents(agents...),
	)

	sess := &clearcast.Session{
		RootAgentID:   config.Session.RootAgentID,
		Mission:       config.Session.Mission,
		Workspace:     make(map[string]any),
		Memory:        make(map[string]any),
		Messages:      []clearcast.Message{},
		CreatedAt:     time.Now(),
		MaxIterations: 10, // Default max iterations
	}

	ctx := context.Background()
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
			log.Fatalf("Runtime Error: %s", e.Error)
		case *clearcast.FinalEvent:
			fmt.Println("\n=== Execution Finished ===")
			fmt.Println(e.Output)
		default:
			// Optionally log unhandled event types
			// log.Printf("Received unhandled event type: %T\n", e)
		}
	}
}
