package handlers

import (
	"fmt"
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
	agent        *services.AgentRunner
	github       *services.GitHubService
	auth         *auth.AuthService
	settings     *services.SettingsService
	clientID     string
	clientSecret string
	redirectURI  string
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(tasks *services.TaskService, events *services.EventBus, agent *services.AgentRunner, db *db.DB) *Handlers {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	githubService := services.NewGitHubService(clientID, clientSecret, redirectURI, db)

	// Initialize auth service
	authService, err := auth.NewAuthServiceFromEnv()
	if err != nil {
		fmt.Printf("Warning: Failed to initialize auth service: %v\n", err)
		authService = nil
	}

	settingsService := services.NewSettingsService(db)

	return &Handlers{
		tasks:        tasks,
		events:       events,
		agent:        agent,
		github:       githubService,
		auth:         authService,
		settings:     settingsService,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

// RegisterRoutes registers all routes on the router.
func (h *Handlers) RegisterRoutes(r chi.Router) {
	// Landing and Auth
	r.Get("/", h.HandleFeed) // New Home is Feed
	r.Get("/auth/login", h.HandleAuth)
	r.Get("/auth/register", h.HandleRegister)
	r.Get("/auth/oauth/{provider}", h.HandleOAuth)
	r.Get("/auth/callback", h.HandleAuthCallback)
	r.Get("/auth/check", h.HandleAuthCheck)
	r.Post("/auth/logout", h.HandleLogout)

	// New UI Routes
	r.Get("/feed", h.HandleFeed)
	r.Get("/feed/active", h.HandleFeedActive)
	r.Get("/task/{id}", h.HandleTaskDetailUI)
	
	// Actions
	r.Post("/add-task", h.HandleAddTask)
	r.Post("/action/retry/{id}", h.HandleActionRetry)
	r.Post("/action/merge/{id}", h.HandleActionMerge)
	r.Post("/action/discard/{id}", h.HandleActionDiscard)
	r.Post("/action/chat/{id}", h.HandleActionChat)
	r.Post("/settings", h.HandleSaveSettings)

	// Keep GitHub OAuth routes
	r.Get("/github/authorize", h.HandleGitHubAuthorize)
	r.Get("/github/callback", h.HandleGitHubCallback)
	r.Post("/disconnect", h.HandleDisconnect)
	
	// Legacy routes redirects
	r.Get("/home", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", http.StatusFound) })
	r.Get("/projects", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", http.StatusFound) })
	r.Get("/board", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/", http.StatusFound) })
}

// maskSensitive masks sensitive values for logging
func maskSensitive(val string) string {
	if val == "" {
		return "not set"
	}
	if len(val) <= 8 {
		return "***"
	}
	return val[:4] + "..." + val[len(val)-4:]
}
