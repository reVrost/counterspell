package llm

// Models are Available models with provider prefix
var Models = []Model{
	// OpenRouter models
	{
		ID:       "o#anthropic/claude-sonnet-4.5",
		Name:     "Claude Sonnet 4.5",
		Provider: "openrouter",
	},
	{
		ID:       "o#anthropic/claude-opus-4.5",
		Name:     "Claude Opus 4.5",
		Provider: "openrouter",
	},
	{
		ID:       "o#google/gemini-3-pro-preview",
		Name:     "Gemini 3 Pro Preview",
		Provider: "openrouter",
	},
	{
		ID:       "o#google/gemini-3-flash-preview",
		Name:     "Gemini 3 Flash Preview",
		Provider: "openrouter",
	},
	{
		ID:       "o#openai/gpt-5.2",
		Name:     "GPT 5.2",
		Provider: "openrouter",
	},
	{
		ID:       "o#openai/gpt-5.1-codex-max",
		Name:     "GPT 5.1 Codex Max",
		Provider: "openrouter",
	},
	// Z.ai models
	{
		ID:       "zai#glm-4.7",
		Name:     "GLM 4.7",
		Provider: "zai",
	},
}

// Model represents an available model
type Model struct {
	ID       string // Unique ID (e.g., "o#anthropic/claude-sonnet-4.5")
	Name     string
	Provider string
}

// ParseModelID parses model ID and returns provider and model name
func ParseModelID(modelID string) (provider, model string) {
	parts := []string{}
	if len(modelID) > 0 {
		// Split by # to get provider
		if idx := findIndex(modelID, "#"); idx != -1 {
			parts = append(parts, modelID[:idx])
			model = modelID[idx+1:]
		}
	}
	if len(parts) > 0 {
		provider = parts[0]
	}
	return
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
