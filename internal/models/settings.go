package models

import "time"

// Agent backend types
const (
	AgentBackendNative     = "native"      // Go-based agent loop (Counterspell)
	AgentBackendClaudeCode = "claude-code" // Claude Code CLI
)

// UserSettings represents the user's API keys and other preferences.
type UserSettings struct {
	UserID        string    `json:"user_id"`
	OpenRouterKey string    `json:"openrouter_key"`
	ZaiKey        string    `json:"zai_key"`
	AnthropicKey  string    `json:"anthropic_key"`
	OpenAIKey     string    `json:"openai_key"`
	AgentBackend  string    `json:"agent_backend"` // "native" or "claude-code"
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetAgentBackend returns the agent backend, defaulting to native.
func (s *UserSettings) GetAgentBackend() string {
	if s.AgentBackend == "" {
		return AgentBackendNative
	}
	return s.AgentBackend
}
