package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/models"
)

// HandleAPITasks returns tasks data as JSON
func (h *Handlers) HandleAPITasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	// Get tasks from DB
	dbTasks, err := h.taskService.List(ctx, userID, nil, nil)
	if err != nil {
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	// Get projects from DB
	dbProjects, err := h.githubService.GetProjects(ctx, userID)
	if err != nil {
		dbProjects = []models.Project{}
	}

	// Convert projects to map for frontend
	projectMap := make(map[string]any)
	for _, p := range dbProjects {
		projectMap[p.ID] = map[string]string{
			"id":    p.ID,
			"name":  p.GitHubOwner + "/" + p.GitHubRepo,
			"icon":  "fa-github",
			"color": "text-blue-400",
		}
	}

	// Organize tasks into categories
	feed := struct {
		Active   []models.Task  `json:"active"`
		Reviews  []models.Task  `json:"reviews"`
		Done     []models.Task  `json:"done"`
		Projects map[string]any `json:"projects"`
	}{
		Active:   []models.Task{},
		Reviews:  []models.Task{},
		Done:     []models.Task{},
		Projects: projectMap,
	}

	for _, t := range dbTasks {
		switch t.Status {
		case models.StatusTodo, models.StatusInProgress:
			feed.Active = append(feed.Active, t)
		case models.StatusReview:
			feed.Reviews = append(feed.Reviews, t)
		case models.StatusDone:
			feed.Done = append(feed.Done, t)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(feed); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// HandleGitHubRepos returns all cached GitHub repositories
func (h *Handlers) HandleGitHubRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	repos, err := h.repoCache.GetCachedRepos(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load repos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(repos); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// HandleSyncRepos manually triggers a sync of GitHub repos to both cache and projects table
func (h *Handlers) HandleSyncRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	// Get GitHub connection
	conn, err := h.githubService.GetActiveConnection(ctx, userID)
	if err != nil || conn == nil {
		http.Error(w, "No GitHub connection found", http.StatusBadRequest)
		return
	}

	// Sync to cache
	if err := h.repoCache.SyncReposFromGitHub(ctx, userID, conn.Token); err != nil {
		slog.Error("[SyncRepos] Failed to sync to cache", "error", err)
	}

	// Sync to projects table
	if err := h.githubService.FetchAndSaveRepositories(ctx, userID, conn); err != nil {
		slog.Error("[SyncRepos] Failed to save projects", "error", err)
		http.Error(w, "Failed to sync repositories", http.StatusInternalServerError)
		return
	}

	slog.Info("[SyncRepos] Sync completed", "user_id", userID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleActivateProject adds a repo to projects list
func (h *Handlers) HandleActivateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	owner := r.FormValue("owner")
	repo := r.FormValue("repo")

	if owner == "" || repo == "" {
		http.Error(w, "Owner and Repo are required", http.StatusBadRequest)
		return
	}

	// Save project to DB
	if err := h.githubService.SaveProject(ctx, userID, owner, repo); err != nil {
		http.Error(w, "Failed to save project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleAPITask returns a single task as JSON
func (h *Handlers) HandleAPITask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	// Get task from DB
	dbTask, err := h.taskService.Get(ctx, userID, taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dbTask); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// HandleAPISession returns current session info as JSON
func (h *Handlers) HandleAPISession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	response := struct {
		Authenticated   bool   `json:"authenticated"`
		Email           string `json:"email,omitempty"`
		GitHubConnected bool   `json:"githubConnected"`
		GitHubLogin     string `json:"githubLogin,omitempty"`
		NeedsGitHubAuth bool   `json:"needsGitHubAuth"`
	}{
		Authenticated:   false,
		GitHubConnected: false,
		NeedsGitHubAuth: false,
	}

	// Get user email from JWT claims (Supabase identity)
	if claims := auth.ClaimsFromContext(ctx); claims != nil {
		response.Email = claims.Email
		response.Authenticated = true
	}

	// Check GitHub connection and validate token
	conn, connErr := h.githubService.GetActiveConnection(ctx, userID)
	if connErr == nil && conn != nil {
		// Validate the stored token is still working
		valid := h.githubService.ValidateToken(ctx, conn.Token)
		if valid {
			response.GitHubConnected = true
			response.GitHubLogin = conn.Login
		} else {
			// Token expired or revoked - need to re-auth
			response.NeedsGitHubAuth = true
			slog.Warn("GitHub token invalid, needs re-auth", "user_id", userID)
		}
	} else if response.Authenticated {
		// User is authenticated via Supabase but has no GitHub connection
		response.NeedsGitHubAuth = true
	}

	// For backwards compatibility: authenticated if either Supabase auth OR valid GitHub
	if response.GitHubConnected {
		response.Authenticated = true
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// HandleAPISettings returns user settings as JSON
func (h *Handlers) HandleAPISettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	settings, err := h.settingsService.GetSettings(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
