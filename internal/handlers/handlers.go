package handlers

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/services"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	tasks        *services.TaskService
	events       *services.EventBus
	orchestrator *services.Orchestrator
	github       *services.GitHubService
	auth         *auth.AuthService
	settings     *services.SettingsService
	clientID     string
	clientSecret string
	redirectURI  string
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(tasks *services.TaskService, events *services.EventBus, database *db.DB, dataDir string) (*Handlers, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	githubService := services.NewGitHubService(clientID, clientSecret, redirectURI, database)
	settingsService := services.NewSettingsService(database)

	// Create orchestrator
	orchestrator, err := services.NewOrchestrator(tasks, githubService, events, settingsService, dataDir)
	if err != nil {
		return nil, err
	}

	// Initialize auth service
	// authService, err := auth.NewAuthServiceFromEnv()
	// if err != nil {
	// 	fmt.Printf("Warning: Failed to initialize auth service: %v\n", err)
	// 	authService = nil
	// }

	return &Handlers{
		tasks:        tasks,
		events:       events,
		orchestrator: orchestrator,
		github:       githubService,
		// Disable for now
		auth:         nil,
		settings:     settingsService,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}, nil
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
	r.Post("/action/merge/{id}", h.HandleActionMerge)
	r.Post("/action/pr/{id}", h.HandleActionPR)
	r.Post("/action/discard/{id}", h.HandleActionDiscard)
	r.Post("/action/chat/{id}", h.HandleActionChat)
	r.Post("/action/resolve-conflict/{id}", h.HandleResolveConflict)
	r.Post("/action/abort-merge/{id}", h.HandleAbortMerge)
	r.Post("/action/complete-merge/{id}", h.HandleCompleteMerge)
	r.Post("/settings", h.HandleSaveSettings)

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
	if h.orchestrator != nil {
		h.orchestrator.Shutdown()
	}
}
