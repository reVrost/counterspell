package services

import (
	"context"
	"testing"

	"github.com/revrost/counterspell/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *db.DB {
	ctx := context.Background()
	testDB, err := db.Connect(ctx, ":memory:")
	require.NoError(t, err, "Failed to create test database")

	err = testDB.RunMigrations(ctx)
	require.NoError(t, err, "Failed to run migrations")

	return testDB
}

func strPtr(s string) *string {
	return &s
}

func TestGetAPIKeyForProvider(t *testing.T) {
	tests := []struct {
		name           string
		settings       *Settings
		requestedProv  string
		expectedKey    string
		expectedProv   string
		expectedModel  string
		expectedErr    bool
	}{
		{
			name: "uses default anthropic when no settings provider and empty requested provider",
			settings: &Settings{
				AnthropicKey: "sk-ant-test",
				AgentBackend:  "native",
			},
			requestedProv: "",
			expectedKey:   "sk-ant-test",
			expectedProv:  "anthropic",
			expectedModel: "claude-opus-4-5",
		},
		{
			name: "uses zai provider when explicitly requested",
			settings: &Settings{
				AnthropicKey: "sk-ant-test",
				ZaiKey:       "zai-test-key",
				AgentBackend:  "native",
			},
			requestedProv: "zai",
			expectedKey:   "zai-test-key",
			expectedProv:  "zai",
			expectedModel: "claude-opus-4-5",
		},
		{
			name: "uses openrouter provider when explicitly requested",
			settings: &Settings{
				AnthropicKey:  "sk-ant-test",
				OpenRouterKey: "or-test-key",
				AgentBackend:   "native",
			},
			requestedProv: "openrouter",
			expectedKey:   "or-test-key",
			expectedProv:  "openrouter",
			expectedModel: "claude-opus-4-5",
		},
		{
			name: "uses openai provider when explicitly requested",
			settings: &Settings{
				AnthropicKey: "sk-ant-test",
				OpenAIKey:     "sk-openai-test",
				AgentBackend:  "native",
			},
			requestedProv: "openai",
			expectedKey:   "sk-openai-test",
			expectedProv:  "openai",
			expectedModel: "claude-opus-4-5",
		},
		{
			name: "uses settings provider when provider param is empty",
			settings: &Settings{
				AnthropicKey: "sk-ant-test",
				ZaiKey:       "zai-test-key",
				Provider:     strPtr("zai"),
				AgentBackend: "native",
			},
			requestedProv: "",
			expectedKey:   "zai-test-key",
			expectedProv:  "zai",
			expectedModel: "claude-opus-4-5",
		},
		{
			name: "uses custom model from settings",
			settings: &Settings{
				AnthropicKey: "sk-ant-test",
				Model:         strPtr("custom-model-v1"),
				AgentBackend: "native",
			},
			requestedProv: "",
			expectedKey:   "sk-ant-test",
			expectedProv:  "anthropic",
			expectedModel: "custom-model-v1",
		},
		{
			name: "returns error for unknown provider",
			settings: &Settings{
				AnthropicKey: "sk-ant-test",
				AgentBackend: "native",
			},
			requestedProv: "unknown-provider",
			expectedErr:   true,
		},
		{
			name: "requested provider overrides settings provider",
			settings: &Settings{
				AnthropicKey: "sk-ant-test",
				ZaiKey:       "zai-test-key",
				Provider:     strPtr("anthropic"),
				AgentBackend: "native",
			},
			requestedProv: "zai",
			expectedKey:   "zai-test-key",
			expectedProv:  "zai",
			expectedModel: "claude-opus-4-5",
		},
		{
			name: "handles all providers with API keys",
			settings: &Settings{
				AnthropicKey:  "sk-ant-test",
				ZaiKey:        "zai-test-key",
				OpenRouterKey: "or-test-key",
				OpenAIKey:     "sk-openai-test",
				AgentBackend:  "native",
			},
			requestedProv: "anthropic",
			expectedKey:   "sk-ant-test",
			expectedProv:  "anthropic",
			expectedModel: "claude-opus-4-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := setupTestDB(t)
			defer testDB.Close()

			svc := NewSettingsService(testDB)
			ctx := context.Background()

			// Insert test settings
			err := svc.UpdateSettings(ctx, tt.settings)
			require.NoError(t, err, "Failed to insert test settings")

			// Call GetAPIKeyForProvider
			key, prov, model, err := svc.GetAPIKeyForProvider(ctx, tt.requestedProv)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, key)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedKey, key, "API key mismatch")
			assert.Equal(t, tt.expectedProv, prov, "Provider mismatch")
			assert.Equal(t, tt.expectedModel, model, "Model mismatch")
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name          string
		settings      *Settings
		expectedKey   string
		expectedProv  string
		expectedModel string
	}{
		{
			name: "returns default anthropic when no settings provider",
			settings: &Settings{
				AnthropicKey: "sk-ant-test",
				AgentBackend: "native",
			},
			expectedKey:   "sk-ant-test",
			expectedProv:  "anthropic",
			expectedModel: "claude-opus-4-5",
		},
		{
			name: "uses settings provider",
			settings: &Settings{
				ZaiKey:       "zai-test-key",
				Provider:     strPtr("zai"),
				AgentBackend: "native",
			},
			expectedKey:   "zai-test-key",
			expectedProv:  "zai",
			expectedModel: "claude-opus-4-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := setupTestDB(t)
			defer testDB.Close()

			svc := NewSettingsService(testDB)
			ctx := context.Background()

			err := svc.UpdateSettings(ctx, tt.settings)
			require.NoError(t, err)

			key, prov, model, err := svc.GetAPIKey(ctx)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedKey, key)
			assert.Equal(t, tt.expectedProv, prov)
			assert.Equal(t, tt.expectedModel, model)
		})
	}
}
