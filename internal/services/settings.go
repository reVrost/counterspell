package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/revrost/counterspell/internal/db"
	"github.com/revrost/counterspell/internal/db/sqlc"
)

// SettingsService handles settings.
type SettingsService struct {
	db *db.DB
}

// Settings represents application settings.
type Settings struct {
	OpenRouterKey string    `json:"openrouter_key,omitempty"`
	ZaiKey        string    `json:"zai_key,omitempty"`
	AnthropicKey  string    `json:"anthropic_key,omitempty"`
	OpenAIKey     string    `json:"openai_key,omitempty"`
	AgentBackend  string    `json:"agent_backend"` // "native", "claude-code", "codex"
	Provider      *string   `json:"provider"`      // "anthropic", "openrouter", etc.
	Model         *string   `json:"model"`         // "claude-opus-4-5", etc.
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewSettingsService creates a new Settings service.
func NewSettingsService(db *db.DB) *SettingsService {
	return &SettingsService{db: db}
}

// GetSettings retrieves settings.
func (s *SettingsService) GetSettings(ctx context.Context) (*Settings, error) {
	row, err := s.db.Queries.GetSettings(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	return &Settings{
		OpenRouterKey: row.OpenrouterKey.String,
		ZaiKey:        row.ZaiKey.String,
		AnthropicKey:  row.AnthropicKey.String,
		OpenAIKey:     row.OpenaiKey.String,
		AgentBackend:  row.AgentBackend,
		Provider:      &row.Provider,
		Model:         &row.Model,
		UpdatedAt:     time.UnixMilli(row.UpdatedAt),
	}, nil
}

// UpdateSettings updates settings with validation.
func (s *SettingsService) UpdateSettings(ctx context.Context, settings *Settings) error {
	// Validate settings
	if err := s.ValidateSettings(settings); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	provider := ""
	if settings.Provider != nil {
		provider = *settings.Provider
	}

	model := ""
	if settings.Model != nil {
		model = *settings.Model
	}

	slog.Info("upserting settings", slog.String("provider", provider), slog.String("model", model), "settings", settings)

	err := s.db.Queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		OpenrouterKey: sql.NullString{String: settings.OpenRouterKey, Valid: settings.OpenRouterKey != ""},
		ZaiKey:        sql.NullString{String: settings.ZaiKey, Valid: settings.ZaiKey != ""},
		AnthropicKey:  sql.NullString{String: settings.AnthropicKey, Valid: settings.AnthropicKey != ""},
		OpenaiKey:     sql.NullString{String: settings.OpenAIKey, Valid: settings.OpenAIKey != ""},
		AgentBackend:  settings.AgentBackend,
		Provider:      sql.NullString{String: provider, Valid: provider != ""},
		Model:         sql.NullString{String: model, Valid: model != ""},
		UpdatedAt:     time.Now().UnixMilli(),
	})
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}
	return nil
}

// ValidateSettings validates settings values.
func (s *SettingsService) ValidateSettings(settings *Settings) error {
	// Validate agent backend
	validBackends := []string{"native", "claude-code", "codex"}
	if settings.AgentBackend == "" {
		return fmt.Errorf("agent_backend is required")
	}
	if !slices.Contains(validBackends, settings.AgentBackend) {
		return fmt.Errorf("invalid agent_backend: %s (must be one of: %s)", settings.AgentBackend, strings.Join(validBackends, ", "))
	}

	// Validate provider
	validProviders := []string{"anthropic", "openrouter", "openai", "zai"}
	if settings.Provider != nil && !slices.Contains(validProviders, *settings.Provider) {
		return fmt.Errorf("invalid provider: %s (must be one of: %s)", *settings.Provider, strings.Join(validProviders, ", "))
	}

	// Validate that selected backend has a corresponding API key
	// provider := "anthropic"
	// if settings.Provider != nil {
	// 	provider = *settings.Provider
	// }

	// switch provider {
	// case "openai":
	// 	if settings.OpenAIKey == "" {
	// 		return fmt.Errorf("OpenAI API key required when using OpenAI provider")
	// 	}
	// case "anthropic":
	// 	if settings.AnthropicKey == "" {
	// 		return fmt.Errorf("anthropic API key required when using Anthropic provider")
	// 	}
	// case "openrouter":
	// 	if settings.OpenRouterKey == "" {
	// 		return fmt.Errorf("OpenRouter API key required when using OpenRouter provider")
	// 	}
	// case "zai":
	// 	if settings.ZaiKey == "" {
	// 		return fmt.Errorf("zai API key required when using Zai provider")
	// 	}
	// case "":
	// No provider set - use default
	// return nil
	// }

	return nil
}

// GetAPIKey returns API key for the current provider.
func (s *SettingsService) GetAPIKey(ctx context.Context) (string, string, string, error) {
	return s.GetAPIKeyForProvider(ctx, "")
}

// GetAPIKeyForProvider returns API key for the specified provider (or default if empty).
func (s *SettingsService) GetAPIKeyForProvider(ctx context.Context, provider string) (string, string, string, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return "", "", "", err
	}
	if settings == nil {
		return "", "", "", errors.New("settings not configured")
	}

	// Use provided provider or default from settings
	if provider == "" {
		provider = "anthropic" // default
		if settings.Provider != nil {
			provider = *settings.Provider
		}
	}

	model := "claude-opus-4-5" // default
	if settings.Model != nil {
		model = *settings.Model
	}

	switch provider {
	case "openai":
		return settings.OpenAIKey, "openai", model, nil
	case "anthropic":
		return settings.AnthropicKey, "anthropic", model, nil
	case "openrouter":
		return settings.OpenRouterKey, "openrouter", model, nil
	case "zai":
		return settings.ZaiKey, "zai", model, nil
	default:
		return "", "", "", fmt.Errorf("unknown provider: %s", provider)
	}
}
