package handlers

import (
	"github.com/revrost/code/counterspell/internal/config"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/services"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	db            *db.DB
	events        *services.EventBus
	cfg           *config.Config
	transcription *services.TranscriptionService

	// Shared services (created once at startup)
	taskService     *services.TaskService
	settingsService *services.SettingsService
	repoManager    *services.RepoManager
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(database *db.DB, events *services.EventBus, cfg *config.Config) (*Handlers, error) {
	transcriptionService := services.NewTranscriptionService()

	return &Handlers{
		db:            database,
		events:        events,
		transcription: transcriptionService,
		cfg:           cfg,

		// Create shared services
		taskService:     services.NewTaskService(database),
		settingsService: services.NewSettingsService(database),
		repoManager:    services.NewRepoManager(cfg.DataDir),
	}, nil
}

// getOrchestrator creates an orchestrator for a task execution.
func (h *Handlers) getOrchestrator(machineID string) (*services.Orchestrator, error) {
	return services.NewOrchestrator(
		h.db,
		h.taskService,
		h.repoManager,
		h.events,
		h.settingsService,
		h.cfg.DataDir,
		machineID,
	)
}
