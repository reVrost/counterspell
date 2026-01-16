package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/models"
)

// SettingsService handles user settings.
type SettingsService struct {
	db *db.DB
}

// NewSettingsService creates a new Settings service.
func NewSettingsService(db *db.DB) *SettingsService {
	return &SettingsService{db: db}
}

// GetSettings retrieves the user settings.
func (s *SettingsService) GetSettings(ctx context.Context) (*models.UserSettings, error) {
	// We assume a single user for now with ID "default"
	query := `SELECT user_id, openrouter_key, zai_key, anthropic_key, openai_key, 
	          COALESCE(agent_backend, 'native') as agent_backend, updated_at 
	          FROM user_settings WHERE user_id = 'default'`

	var settings models.UserSettings
	err := s.db.QueryRowContext(ctx, query).Scan(
		&settings.UserID,
		&settings.OpenRouterKey,
		&settings.ZaiKey,
		&settings.AnthropicKey,
		&settings.OpenAIKey,
		&settings.AgentBackend,
		&settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return empty settings with default backend
		return &models.UserSettings{UserID: "default", AgentBackend: models.AgentBackendNative}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	return &settings, nil
}

// UpdateSettings updates the user settings.
func (s *SettingsService) UpdateSettings(ctx context.Context, settings *models.UserSettings) error {
	query := `INSERT INTO user_settings (user_id, openrouter_key, zai_key, anthropic_key, openai_key, agent_backend, updated_at)
	          VALUES ('default', ?, ?, ?, ?, ?, ?)
	          ON CONFLICT(user_id) DO UPDATE SET
	          openrouter_key = excluded.openrouter_key,
	          zai_key = excluded.zai_key,
	          anthropic_key = excluded.anthropic_key,
	          openai_key = excluded.openai_key,
	          agent_backend = excluded.agent_backend,
	          updated_at = excluded.updated_at`

	_, err := s.db.ExecContext(ctx, query,
		settings.OpenRouterKey,
		settings.ZaiKey,
		settings.AnthropicKey,
		settings.OpenAIKey,
		settings.GetAgentBackend(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}
