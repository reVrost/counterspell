package llm

// Provider defines the interface for LLM providers.
type Provider interface {
	// APIURL returns the base URL for the API.
	APIURL() string

	// APIVersion returns the API version header value (if applicable).
	APIVersion() string

	// APIKey returns the API key.
	APIKey() string

	// Model returns the default model to use.
	Model() string

	// SetModel sets the model to use.
	SetModel(model string)

	// Type returns the provider type (anthropic or openai).
	Type() string
}

// AnthropicProvider implements Anthropic API.
type AnthropicProvider struct {
	apiKey string
	model  string
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  "claude-opus-4-5",
	}
}

func (p *AnthropicProvider) Type() string {
	return "anthropic"
}

func (p *AnthropicProvider) APIURL() string {
	return "https://api.anthropic.com/v1/messages"
}

func (p *AnthropicProvider) APIVersion() string {
	return "2023-06-01"
}

func (p *AnthropicProvider) APIKey() string {
	return p.apiKey
}

func (p *AnthropicProvider) Model() string {
	return p.model
}

func (p *AnthropicProvider) SetModel(model string) {
	p.model = model
}

// OpenRouterProvider implements OpenRouter API.
type OpenRouterProvider struct {
	apiKey string
	model  string
}

func NewOpenRouterProvider(apiKey string) *OpenRouterProvider {
	return &OpenRouterProvider{
		apiKey: apiKey,
		model:  "anthropic/claude-sonnet-4.5",
	}
}

func (p *OpenRouterProvider) Type() string {
	return "anthropic"
}

func (p *OpenRouterProvider) APIURL() string {
	return "https://openrouter.ai/api/v1/messages"
}

func (p *OpenRouterProvider) APIVersion() string {
	return ""
}

func (p *OpenRouterProvider) APIKey() string {
	return p.apiKey
}

func (p *OpenRouterProvider) Model() string {
	return p.model
}

func (p *OpenRouterProvider) SetModel(model string) {
	p.model = model
}

// ZaiProvider implements Z.ai API.
type ZaiProvider struct {
	apiKey string
	model  string
}

func NewZaiProvider(apiKey string) *ZaiProvider {
	return &ZaiProvider{
		apiKey: apiKey,
		model:  "glm-4.7",
	}
}

func (p *ZaiProvider) Type() string {
	return "openai"
}

func (p *ZaiProvider) APIURL() string {
	return "https://api.z.ai/api/coding/paas/v4/chat/completions"
}

func (p *ZaiProvider) APIVersion() string {
	return ""
}

func (p *ZaiProvider) APIKey() string {
	return p.apiKey
}

func (p *ZaiProvider) Model() string {
	return p.model
}

func (p *ZaiProvider) SetModel(model string) {
	p.model = model
}
