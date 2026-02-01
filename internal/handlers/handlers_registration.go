package handlers

import (
	"log/slog"
	"sync"

	"github.com/revrost/counterspell/internal/config"
	"github.com/revrost/counterspell/internal/db"
	"github.com/revrost/counterspell/internal/services"
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
	oauthService    *services.OAuthService
	gitReposManager *services.GitManager

	// Track active orchestrators for shutdown
	orchestrators map[string]*services.Orchestrator
	mu            sync.Mutex
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
		oauthService:    services.NewOAuthService(database, cfg),
		gitReposManager: services.NewGitManager(cfg.DataDir),

		// Initialize orchestrator tracking
		orchestrators: make(map[string]*services.Orchestrator),
	}, nil
}

// getOrchestrator creates an orchestrator for a task execution.
// For local-first single-tenant mode, we use a fixed userID "default".
// We create a single shared orchestrator for all tasks.
func (h *Handlers) getOrchestrator() (*services.Orchestrator, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Use a single shared orchestrator for "shared"
	if orch, ok := h.orchestrators["shared"]; ok {
		return orch, nil
	}

	// Create the shared orchestrator
	orch, err := services.NewOrchestrator(
		h.taskService,
		h.events,
		h.settingsService,
		h.githubService,
		h.cfg.DataDir,
	)
	if err != nil {
		return nil, err
	}

	h.orchestrators["shared"] = orch
	return orch, nil
}

// Shutdown gracefully shuts down all active orchestrators.
func (h *Handlers) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for name, orch := range h.orchestrators {
		slog.Info("[HANDLERS] Shutting down orchestrator", "name", name)
		orch.Shutdown()
	}
	slog.Info("[HANDLERS] All orchestrators shut down")
}
