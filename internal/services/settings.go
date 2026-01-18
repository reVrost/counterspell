package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
func (s *SettingsService) GetSettings(ctx context.Context, userID string) (*models.UserSettings, error) {
	row, err := s.db.Queries.GetUserSettings(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Return empty settings with default backend
		return &models.UserSettings{UserID: userID, AgentBackend: models.AgentBackendNative}, nil
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
func (s *SettingsService) UpdateSettings(ctx context.Context, userID string, settings *models.UserSettings) error {
	err := s.db.Queries.UpsertUserSettings(ctx, sqlc.UpsertUserSettingsParams{
		UserID:        userID,
		OpenrouterKey: pgtype.Text{String: settings.OpenRouterKey, Valid: settings.OpenRouterKey != ""},
		ZaiKey:        pgtype.Text{String: settings.ZaiKey, Valid: settings.ZaiKey != ""},
		AnthropicKey:  pgtype.Text{String: settings.AnthropicKey, Valid: settings.AnthropicKey != ""},
		OpenaiKey:     pgtype.Text{String: settings.OpenAIKey, Valid: settings.OpenAIKey != ""},
		AgentBackend:  pgtype.Text{String: settings.GetAgentBackend(), Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}
