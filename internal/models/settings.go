package models

import "time"

// UserSettings represents the user's API keys and other preferences.
type UserSettings struct {
	UserID       string    `json:"user_id"`
	OpenRouterKey string    `json:"openrouter_key"`
	ZaiKey        string    `json:"zai_key"`
	AnthropicKey  string    `json:"anthropic_key"`
	OpenAIKey     string    `json:"openai_key"`
	UpdatedAt     time.Time `json:"updated_at"`
}
