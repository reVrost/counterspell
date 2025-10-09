package clearcast

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"maps"

	"github.com/goccy/go-yaml"
	"github.com/revrost/go-openrouter"
)

// AgentDef defines the structure for an agent in the YAML configuration.
type AgentDef struct {
	ID       string    `yaml:"id"`
	Mode     string    `yaml:"mode"`
	Model    string    `yaml:"model"`
	Prompt   string    `yaml:"prompt"`
	Toolsets []Toolset `yaml:"toolsets,omitempty"` // Tools specific to this agent
}

type Toolset struct {
	Type        string         `yaml:"type"`
	ID          string         `yaml:"id"`
	Remote      *RemoteToolset `yaml:"remote,omitempty"`
	Prompt      string         `yaml:"prompt,omitempty"`
	Description string         `yaml:"description,omitempty"`
	Model       string         `yaml:"model,omitempty"`
}

type RemoteToolset struct {
	URL           string `yaml:"url"`
	TransportType string `yaml:"transport_type"`
}

// SessionDef defines the structure for the session in the YAML configuration.
type SessionDef struct {
	RootAgentID string         `yaml:"root_agent_id"`
	Extra       map[string]any `yaml:",inline"` // Captures all other fields like topic, etc.
	// Mission     string         `yaml:"mission"`
}

// OrchestrationFile defines the top-level structure of the YAML file.
type OrchestrationFile struct {
	Agents  []AgentDef `yaml:"agents"`
	Session SessionDef `yaml:"session"`
}

// NewRunFromPath creates a new runtime from a config file (yaml)
func NewRunFromPath(yamlFilepath string) (Runtime, error) {

	data, err := os.ReadFile(yamlFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", yamlFilepath, err)
	}

	var config OrchestrationFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	slog.Info("Loaded YAML config", "file", yamlFilepath, "config", config)

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY environment variable not set")
	}
	orClient := openrouter.NewClient(apiKey)
	llmProvider := NewOpenRouterProvider(orClient)

	agents := make([]*Agent, 0, len(config.Agents))
	for _, agentDef := range config.Agents {

		opts := []AgentOption{}
		for _, toolDef := range agentDef.Toolsets {
			switch toolDef.Type {
			case "call":
				opts = append(opts, WithDefaultTools(toolDef.ID))
			case "llm":
				opts = append(opts, WithTool(&Tool{
					ID: toolDef.ID,
					Execute: func(ctx context.Context, params map[string]any) (any, error) {
						result, err := llmProvider.ChatCompletion(ctx, ChatCompletionRequest{
							Model:    toolDef.Model,
							Messages: []ChatMessage{SystemMessage(toolDef.Prompt)},
						})
						if err != nil {
							return nil, err
						}
						return result, nil
					},
					Description: toolDef.Description,
				}))

			case "mcp":
				// TODO: implement MCP tool
			}
		}
		agent := NewAgent(agentDef.ID, agentDef.Model, agentDef.Prompt, llmProvider, opts...)
		agents = append(agents, agent)
	}

	// Initialize workspace with session extra fields (like topic)
	workspace := make(map[string]any)

	maps.Copy(workspace, config.Session.Extra)
	slog.Info("Initialized workspace", "workspace", workspace)

	sess := &Session{
		RootAgentID:   config.Session.RootAgentID,
		Workspace:     workspace,
		Memory:        make(map[string]any),
		Messages:      []Message{},
		CreatedAt:     time.Now(),
		MaxIterations: 10, // Default max iterations
	}

	rt := NewRuntime(
		WithAgents(agents...),
		WithSession(sess),
	)
	return rt, nil
}
