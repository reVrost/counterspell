package handlers

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
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

// RegisterRoutes registers all routes on the router.
func (h *Handlers) RegisterRoutes(r chi.Router) {
	// Landing and Auth
	r.Get("/", h.HandleFeed)
	r.Get("/auth/login", h.HandleAuth)
	r.Get("/auth/register", h.HandleRegister)
	r.Get("/auth/oauth/{provider}", h.HandleOAuth)
	r.Get("/auth/callback", h.HandleAuthCallback)
	r.Get("/auth/check", h.HandleAuthCheck)
	r.Post("/auth/logout", h.HandleLogout)

	// Feed Routes
	r.Get("/feed", h.HandleFeed)
	r.Get("/feed/active", h.HandleFeedActive)
	r.Get("/feed/stream", h.HandleFeedActiveSSE)
	r.Get("/task/{id}", h.HandleTaskDetailUI)

	// SSE for streaming
	r.Get("/events", h.HandleSSE)
	r.Get("/task/{id}/stream", h.HandleTaskSSE) // Unified stream for task updates
	// Deprecated: individual streams (kept for backwards compatibility)
	r.Get("/task/{id}/logs/stream", h.HandleTaskLogsSSE)
	r.Get("/task/{id}/diff/stream", h.HandleTaskDiffSSE)
	r.Get("/task/{id}/agent/stream", h.HandleTaskAgentSSE)

	// Actions
	r.Post("/add-task", h.HandleAddTask)
	r.Post("/action/retry/{id}", h.HandleActionRetry)
	r.Post("/action/clear/{id}", h.HandleActionClear)
	r.Post("/action/merge/{id}", h.HandleActionMerge)
	r.Post("/action/pr/{id}", h.HandleActionPR)
	r.Post("/action/discard/{id}", h.HandleActionDiscard)
	r.Post("/action/chat/{id}", h.HandleActionChat)
	r.Post("/action/resolve-conflict/{id}", h.HandleResolveConflict)
	r.Post("/action/abort-merge/{id}", h.HandleAbortMerge)
	r.Post("/action/complete-merge/{id}", h.HandleCompleteMerge)
	r.Post("/settings", h.HandleSaveSettings)
	r.Post("/transcribe", h.HandleTranscribe)
	r.Get("/api/files/search", h.HandleFileSearch)

	// GitHub OAuth routes
	r.Get("/github/authorize", h.HandleGitHubAuthorize)
	r.Get("/github/callback", h.HandleGitHubCallback)
	r.Post("/disconnect", h.HandleDisconnect)

	// Legacy routes redirects
	r.Get("/home", redirectTo("/"))
	r.Get("/projects", redirectTo("/"))
	r.Get("/board", redirectTo("/"))
}

func redirectTo(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path, http.StatusFound)
	}
}

// Shutdown gracefully shuts down the handlers.
func (h *Handlers) Shutdown() {
	// Registry handles cleanup of user managers
}
