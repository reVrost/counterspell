package agent

import (
	"testing"

	"github.com/revrost/counterspell/internal/llm"
)

type callerProviderStub struct {
	providerType string
}

func (p *callerProviderStub) APIURL() string        { return "" }
func (p *callerProviderStub) APIVersion() string    { return "" }
func (p *callerProviderStub) APIKey() string        { return "" }
func (p *callerProviderStub) Model() string         { return "test" }
func (p *callerProviderStub) SetModel(model string) {}
func (p *callerProviderStub) Type() string          { return p.providerType }

var _ llm.Provider = (*callerProviderStub)(nil)

func TestNewLLMCaller_SelectsByProviderType(t *testing.T) {
	tests := []struct {
		name         string
		providerType string
		wantType     any
	}{
		{name: "anthropic", providerType: "anthropic", wantType: &AnthropicCaller{}},
		{name: "openai", providerType: "openai", wantType: &OpenAICaller{}},
		{name: "openrouter", providerType: "openrouter", wantType: &OpenRouterCaller{}},
		{name: "zai-coding", providerType: "zai-coding", wantType: &ZAICodingCaller{}},
		{name: "fallback", providerType: "unknown", wantType: &AnthropicCaller{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			caller := NewLLMCaller(&callerProviderStub{providerType: tc.providerType})
			switch tc.wantType.(type) {
			case *AnthropicCaller:
				if _, ok := caller.(*AnthropicCaller); !ok {
					t.Fatalf("expected AnthropicCaller, got %T", caller)
				}
			case *OpenAICaller:
				if _, ok := caller.(*OpenAICaller); !ok {
					t.Fatalf("expected OpenAICaller, got %T", caller)
				}
			case *OpenRouterCaller:
				if _, ok := caller.(*OpenRouterCaller); !ok {
					t.Fatalf("expected OpenRouterCaller, got %T", caller)
				}
			case *ZAICodingCaller:
				if _, ok := caller.(*ZAICodingCaller); !ok {
					t.Fatalf("expected ZAICodingCaller, got %T", caller)
				}
			}
		})
	}
}
