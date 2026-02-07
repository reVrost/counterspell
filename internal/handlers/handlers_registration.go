package handlers

import (
	"github.com/revrost/counterspell/internal/config"
	"github.com/revrost/counterspell/internal/services"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	cfg *config.Config

	repository      *services.Repository
	sessionService  *services.SessionService
	settingsService *services.SettingsService
	fileService     *services.FileService
	oauthService    *services.OAuthService
	openAIConnector *services.OpenAIConnectorService
	orchestrator    *services.Orchestrator
	eventBus        *services.EventBus
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(
	cfg *config.Config,
	repository *services.Repository,
	eventBus *services.EventBus,
	settingsService *services.SettingsService,
	sessionService *services.SessionService,
	oauthService *services.OAuthService,
	openAIConnector *services.OpenAIConnectorService,
	orchestrator *services.Orchestrator) (*Handlers, error) {

	return &Handlers{
		cfg:             cfg,
		repository:      repository,
		eventBus:        eventBus,
		sessionService:  sessionService,
		settingsService: settingsService,
		fileService:     services.NewFileService(cfg.DataDir),
		oauthService:    oauthService,
		openAIConnector: openAIConnector,
		orchestrator:    orchestrator,
	}, nil
}
