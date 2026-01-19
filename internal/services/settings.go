package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
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
	AgentBackend  string    `json:"agent_backend"` // "native", "openai", "anthropic", "openrouter", "zai"
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
	}, nil
}

// UpdateSettings updates settings with validation.
func (s *SettingsService) UpdateSettings(ctx context.Context, settings *Settings) error {
	// Validate settings
	if err := s.ValidateSettings(settings); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	err := s.db.Queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		OpenrouterKey: sql.NullString{String: settings.OpenRouterKey, Valid: settings.OpenRouterKey != ""},
		ZaiKey:        sql.NullString{String: settings.ZaiKey, Valid: settings.ZaiKey != ""},
		AnthropicKey:  sql.NullString{String: settings.AnthropicKey, Valid: settings.AnthropicKey != ""},
		OpenaiKey:     sql.NullString{String: settings.OpenAIKey, Valid: settings.OpenAIKey != ""},
		AgentBackend:  settings.AgentBackend,
	})
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}
	return nil
}

// ValidateSettings validates settings values.
func (s *SettingsService) ValidateSettings(settings *Settings) error {
	// Validate agent backend
	validBackends := []string{"native", "openai", "anthropic", "openrouter", "zai"}
	if settings.AgentBackend == "" {
		return fmt.Errorf("agent_backend is required")
	}
	if !slices.Contains(validBackends, settings.AgentBackend) {
		return fmt.Errorf("invalid agent_backend: %s (must be one of: %s)", settings.AgentBackend, strings.Join(validBackends, ", "))
	}

	// Validate that the selected backend has a corresponding API key
	switch settings.AgentBackend {
	case "openai":
		if settings.OpenAIKey == "" {
			return fmt.Errorf("OpenAI API key required when using OpenAI backend")
		}
	case "anthropic":
		if settings.AnthropicKey == "" {
			return fmt.Errorf("anthropic API key required when using Anthropic backend")
		}
	case "openrouter":
		if settings.OpenRouterKey == "" {
			return fmt.Errorf("OpenRouter API key required when using OpenRouter backend")
		}
	case "zai":
		if settings.ZaiKey == "" {
			return fmt.Errorf("zai API key required when using Zai backend")
		}
	case "native":
		// No API key needed for native mode
	default:
		return fmt.Errorf("unsupported agent_backend: %s", settings.AgentBackend)
	}

	return nil
}

// GetAPIKey returns the API key for the current backend.
func (s *SettingsService) GetAPIKey(ctx context.Context) (string, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return "", err
	}

	switch settings.AgentBackend {
	case "openai":
		return settings.OpenAIKey, nil
	case "anthropic":
		return settings.AnthropicKey, nil
	case "openrouter":
		return settings.OpenRouterKey, nil
	case "zai":
		return settings.ZaiKey, nil
	case "native":
		return "", nil
	default:
		return "", fmt.Errorf("unknown agent_backend: %s", settings.AgentBackend)
	}
}
