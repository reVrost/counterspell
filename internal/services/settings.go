package services

import (
	"context"
	"database/sql"
	"fmt"
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
	OpenRouterKey string
	ZaiKey        string
	AnthropicKey  string
	OpenAIKey     string
	AgentBackend  string
	UpdatedAt     time.Time
}

// NewSettingsService creates a new Settings service.
func NewSettingsService(db *db.DB) *SettingsService {
	return &SettingsService{db: db}
}

// GetSettings retrieves settings.
func (s *SettingsService) GetSettings(ctx context.Context) (*Settings, error) {
	row, err := s.db.Queries.GetSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	return &Settings{
		OpenRouterKey: row.OpenrouterKey.String,
		ZaiKey:        row.ZaiKey.String,
		AnthropicKey:  row.AnthropicKey.String,
		OpenAIKey:     row.OpenaiKey.String,
		AgentBackend:  row.AgentBackend,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

// UpdateSettings updates settings.
func (s *SettingsService) UpdateSettings(ctx context.Context, settings *Settings) error {
	err := s.db.Queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		OpenrouterKey: sql.NullString{String: settings.OpenRouterKey, Valid: settings.OpenRouterKey != ""},
		ZaiKey:        sql.NullString{String: settings.ZaiKey, Valid: settings.ZaiKey != ""},
		AnthropicKey:  sql.NullString{String: settings.AnthropicKey, Valid: settings.AnthropicKey != ""},
		OpenaiKey:     sql.NullString{String: settings.OpenAIKey, Valid: settings.OpenAIKey != ""},
		AgentBackend:  sql.NullString{String: settings.AgentBackend, Valid: settings.AgentBackend != ""},
		UpdatedAt:     sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}
	return nil
}
