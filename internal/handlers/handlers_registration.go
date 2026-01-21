package handlers

import (
	"github.com/revrost/code/counterspell/internal/config"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/services"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	events *services.EventBus
	cfg    *config.Config

	// Shared services (created once at startup)
	transcription   *services.TranscriptionService
	taskService     *services.Repository
	settingsService *services.SettingsService
	fileService     *services.FileService
	githubService   *services.GitHubService
	gitReposManager *services.GitManager
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(database *db.DB, events *services.EventBus, cfg *config.Config) (*Handlers, error) {
	transcriptionService := services.NewTranscriptionService()

	return &Handlers{
		events:        events,
		transcription: transcriptionService,
		cfg:           cfg,

		// Create shared services
		taskService:     services.NewRepository(database),
		settingsService: services.NewSettingsService(database),
		fileService:     services.NewFileService(cfg.DataDir),
		githubService:   services.NewGitHubService(database, cfg.GitHubClientID, cfg.GitHubClientSecret),
		gitReposManager: services.NewGitManager(cfg.DataDir),
	}, nil
}

// getOrchestrator creates an orchestrator for a task execution.
// For local-first single-tenant mode, we use a fixed userID "default".
func (h *Handlers) getOrchestrator() (*services.Orchestrator, error) {
	return services.NewOrchestrator(
		h.taskService,
		h.events,
		h.settingsService,
		h.githubService,
		h.cfg.DataDir,
	)
}
