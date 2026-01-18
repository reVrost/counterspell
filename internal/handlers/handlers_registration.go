package handlers

import (
	"os"

	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/config"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/services"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	db            *db.DB
	events        *services.EventBus
	authService   *auth.AuthService
	transcription *services.TranscriptionService
	cfg           *config.Config
	clientID      string
	clientSecret  string
	redirectURI   string

	// Shared services (created once at startup)
	taskService     *services.TaskService
	githubService   *services.GitHubService
	settingsService *services.SettingsService
	repoCache       *services.RepoCache
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(database *db.DB, events *services.EventBus, cfg *config.Config) (*Handlers, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	transcriptionService := services.NewTranscriptionService()

	// Initialize auth service
	authService, err := auth.NewAuthServiceFromEnv()
	if err != nil {
		// Auth service is optional - just log and continue
		authService = nil
	}

	return &Handlers{
		db:            database,
		events:        events,
		authService:   authService,
		transcription: transcriptionService,
		cfg:           cfg,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   redirectURI,

		// Create shared services
		taskService:     services.NewTaskService(database),
		githubService:   services.NewGitHubService(clientID, clientSecret, redirectURI, database),
		settingsService: services.NewSettingsService(database),
		repoCache:       services.NewRepoCache(database),
	}, nil
}

// getOrchestrator creates an orchestrator for a task execution.
func (h *Handlers) getOrchestrator(userID string) (*services.Orchestrator, error) {
	return services.NewOrchestrator(
		h.taskService,
		h.githubService,
		h.events,
		h.settingsService,
		h.cfg.DataDir,
		userID,
	)
}
