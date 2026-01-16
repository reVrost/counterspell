package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
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
	row, err := s.db.Queries.GetUserSettings(ctx)
	if err == sql.ErrNoRows {
		// Return empty settings with default backend
		return &models.UserSettings{UserID: "default", AgentBackend: models.AgentBackendNative}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	settings := &models.UserSettings{
		UserID:        row.UserID,
		OpenRouterKey: row.OpenrouterKey.String,
		ZaiKey:        row.ZaiKey.String,
		AnthropicKey:  row.AnthropicKey.String,
		OpenAIKey:     row.OpenaiKey.String,
		AgentBackend:  row.AgentBackend,
		UpdatedAt:     row.UpdatedAt.Time,
	}

	return settings, nil
}

// UpdateSettings updates the user settings.
func (s *SettingsService) UpdateSettings(ctx context.Context, settings *models.UserSettings) error {
	err := s.db.Queries.UpsertUserSettings(ctx, sqlc.UpsertUserSettingsParams{
		OpenrouterKey: sql.NullString{String: settings.OpenRouterKey, Valid: settings.OpenRouterKey != ""},
		ZaiKey:        sql.NullString{String: settings.ZaiKey, Valid: settings.ZaiKey != ""},
		AnthropicKey:  sql.NullString{String: settings.AnthropicKey, Valid: settings.AnthropicKey != ""},
		OpenaiKey:     sql.NullString{String: settings.OpenAIKey, Valid: settings.OpenAIKey != ""},
		AgentBackend:  sql.NullString{String: settings.GetAgentBackend(), Valid: true},
		UpdatedAt:     sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}
