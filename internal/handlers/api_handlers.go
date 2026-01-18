package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
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
		_ = render.Render(w, r, ErrInternalServer("Failed to load tasks", err))
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
		case models.StatusPending, models.StatusInProgress:
			feed.Active = append(feed.Active, t)
		case models.StatusReview:
			feed.Reviews = append(feed.Reviews, t)
		case models.StatusDone, models.StatusFailed:
			feed.Done = append(feed.Done, t)
		}
	}

	render.JSON(w, r, feed)
}

// HandleGitHubRepos returns all cached GitHub repositories
func (h *Handlers) HandleGitHubRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	repos, err := h.repoCache.GetCachedRepos(ctx, userID)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Failed to load repos", err))
		return
	}

	render.JSON(w, r, repos)
}

// HandleSyncRepos manually triggers a sync of GitHub repos to both cache and projects table
func (h *Handlers) HandleSyncRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	// Get GitHub connection
	conn, err := h.githubService.GetActiveConnection(ctx, userID)
	if err != nil || conn == nil {
		_ = render.Render(w, r, ErrInvalidRequest(nil))
		return
	}

	// Sync to cache
	if err := h.repoCache.SyncReposFromGitHub(ctx, userID, conn.Token); err != nil {
		slog.Error("[SyncRepos] Failed to sync to cache", "error", err)
	}

	// Sync to projects table
	if err := h.githubService.FetchAndSaveRepositories(ctx, userID, conn); err != nil {
		slog.Error("[SyncRepos] Failed to save projects", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to sync repositories", err))
		return
	}

	slog.Info("[SyncRepos] Sync completed", "user_id", userID)
	_ = render.Render(w, r, Success("Sync completed"))
}

// HandleActivateProject adds a repo to projects list
func (h *Handlers) HandleActivateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	data := &ActivateProjectRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Save project to DB
	if err := h.githubService.SaveProject(ctx, userID, data.Owner, data.Repo); err != nil {
		_ = render.Render(w, r, ErrInternalServer("Failed to save project", err))
		return
	}

	_ = render.Render(w, r, Success("Project activated"))
}

// HandleAPITask returns a single task as JSON with project info
func (h *Handlers) HandleAPITask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	// Get task with project info in single query
	taskWithProject, err := h.taskService.GetWithProject(ctx, userID, taskID)
	if err != nil {
		_ = render.Render(w, r, ErrNotFound("Task not found"))
		return
	}

	// Build project map for frontend
	project := map[string]string{
		"id":    "",
		"name":  "Unknown",
		"icon":  "fa-github",
		"color": "text-blue-400",
	}
	if taskWithProject.Project != nil {
		project["id"] = taskWithProject.Project.ID
		project["name"] = taskWithProject.Project.GitHubOwner + "/" + taskWithProject.Project.GitHubRepo
	}

	// Build response
	response := struct {
		Task     *models.Task      `json:"task"`
		Project  map[string]string `json:"project"`
		Messages []models.Message  `json:"messages"`
		Logs     []any             `json:"logs"`
	}{
		Task:     taskWithProject.Task,
		Project:  project,
		Messages: []models.Message{},
		Logs:     []any{},
	}

	render.JSON(w, r, response)
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

	render.JSON(w, r, response)
}

// HandleAPISettings returns user settings as JSON
func (h *Handlers) HandleAPISettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	settings, err := h.settingsService.GetSettings(ctx, userID)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Failed to load settings", err))
		return
	}

	render.JSON(w, r, settings)
}

// HandleHome returns a JSON object with app metadata
func (h *Handlers) HandleHome(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{
		"name":        "Counterspell",
		"version":     "2.1.0",
		"description": "Mobile-first, hosted AI agent Kanban.",
	})
}

// HandleTaskDetailUI returns task details as JSON (alias to HandleAPITask)
func (h *Handlers) HandleTaskDetailUI(w http.ResponseWriter, r *http.Request) {
	h.HandleAPITask(w, r)
}

// HandleAddTask creates a new task and starts execution.
func (h *Handlers) HandleAddTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	data := &AddTaskRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	slog.Info("Adding task", "intent", data.VoiceInput)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		slog.Error("Failed to get orchestrator", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to get orchestrator", err))
		return
	}

	_, err = orchestrator.StartTask(ctx, data.ProjectID, data.VoiceInput, data.ModelID)
	if err != nil {
		slog.Error("Failed to start task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to start task: "+err.Error(), err))
		return
	}

	render.Status(r, http.StatusCreated)
	_ = render.Render(w, r, Success("Task started"))
}

// HandleSaveSettings saves user settings
func (h *Handlers) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	data := &SaveSettingsRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if userID == "" {
		userID = "default"
	}

	settings := &models.UserSettings{
		UserID:        userID,
		OpenRouterKey: data.OpenRouterKey,
		ZaiKey:        data.ZaiKey,
		AnthropicKey:  data.AnthropicKey,
		OpenAIKey:     data.OpenAIKey,
		AgentBackend:  data.AgentBackend,
	}

	if err := h.settingsService.UpdateSettings(ctx, userID, settings); err != nil {
		_ = render.Render(w, r, ErrInternalServer("Failed to save settings: "+err.Error(), err))
		return
	}

	_ = render.Render(w, r, Success("Settings saved successfully"))
}

// HandleFileSearch searches for files in a project using fuzzy matching.
func (h *Handlers) HandleFileSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)
	projectID := r.URL.Query().Get("project_id")
	query := r.URL.Query().Get("q")

	if projectID == "" {
		_ = render.Render(w, r, ErrInvalidRequest(nil))
		return
	}

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		render.JSON(w, r, []string{})
		return
	}

	files, err := orchestrator.SearchProjectFiles(ctx, projectID, query, 20)
	if err != nil {
		slog.Error("File search failed", "error", err)
		render.JSON(w, r, []string{})
		return
	}

	if files == nil {
		files = []string{}
	}

	render.JSON(w, r, files)
}

// HandleTranscribe transcribes uploaded audio to text.
func (h *Handlers) HandleTranscribe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")

	text, err := h.transcription.TranscribeAudio(ctx, file, contentType)
	if err != nil {
		slog.Error("Transcription failed", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Transcription failed: "+err.Error(), err))
		return
	}

	render.PlainText(w, r, text)
}
