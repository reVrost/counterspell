package handlers

import (
	"github.com/revrost/counterspell/internal/config"
	"github.com/revrost/counterspell/internal/services"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	cfg *config.Config

	sessionService  *services.SessionService
	settingsService *services.SettingsService
	fileService     *services.FileService
	oauthService    *services.OAuthService
	orchestrator    *services.Orchestrator
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(
	cfg *config.Config,
	settingsService *services.SettingsService,
	sessionService *services.SessionService,
	oauthService *services.OAuthService,
	orchestrator *services.Orchestrator) (*Handlers, error) {

	return &Handlers{
		cfg:             cfg,
		sessionService:  sessionService,
		settingsService: settingsService,
		fileService:     services.NewFileService(cfg.DataDir),
		oauthService:    oauthService,
		orchestrator:    orchestrator,
	}, nil
}
