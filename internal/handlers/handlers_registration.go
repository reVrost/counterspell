package handlers

import (
	"context"
	"os"

	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/config"
	"github.com/revrost/code/counterspell/internal/services"
)

// UserServices holds per-user service instances.
type UserServices struct {
	Tasks    *services.TaskService
	GitHub   *services.GitHubService
	Settings *services.SettingsService
}

// Handlers contains all HTTP handlers.
type Handlers struct {
	registry      *services.UserManagerRegistry
	events        *services.EventBus
	auth          *auth.AuthService
	transcription *services.TranscriptionService
	cfg           *config.Config
	clientID      string
	clientSecret  string
	redirectURI   string
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(registry *services.UserManagerRegistry, events *services.EventBus, cfg *config.Config) (*Handlers, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	transcriptionService := services.NewTranscriptionService()

	// Initialize auth service for multi-tenant mode
	var authService *auth.AuthService
	var err error
	if cfg.MultiTenant {
		authService, err = auth.NewAuthServiceFromEnv()
		if err != nil {
			return nil, err
		}
	}

	return &Handlers{
		registry:      registry,
		events:        events,
		auth:          authService,
		transcription: transcriptionService,
		cfg:           cfg,
		clientID:      clientID,
		clientSecret:  clientSecret,
		redirectURI:   redirectURI,
	}, nil
}

// getServices returns per-user services for the current request.
// This is the core of multi-tenant support - each user gets their own DB and services.
func (h *Handlers) getServices(ctx context.Context) (*UserServices, error) {
	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		userID = "default"
	}

	um, err := h.registry.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	db := um.DB()
	return &UserServices{
		Tasks:    services.NewTaskService(db),
		GitHub:   services.NewGitHubService(h.clientID, h.clientSecret, h.redirectURI, db),
		Settings: services.NewSettingsService(db),
	}, nil
}

// getOrchestrator creates an orchestrator for the current user.
// Orchestrators are created per-request because they hold user-specific services.
func (h *Handlers) getOrchestrator(ctx context.Context) (*services.Orchestrator, error) {
	svc, err := h.getServices(ctx)
	if err != nil {
		return nil, err
	}

	return services.NewOrchestrator(svc.Tasks, svc.GitHub, h.events, svc.Settings, h.cfg.DataDir)
}
