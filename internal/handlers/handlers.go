package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/services"
	"github.com/revrost/code/counterspell/internal/ui"
	"github.com/revrost/code/counterspell/internal/utils"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	tasks        *services.TaskService
	events       *services.EventBus
	agent        *services.AgentRunner
	github       *services.GitHubService
	auth         *auth.AuthService
	clientID     string
	clientSecret string
	redirectURI  string
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(tasks *services.TaskService, events *services.EventBus, agent *services.AgentRunner, db *db.DB) *Handlers {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	// Log GitHub configuration (don't log actual secrets)
	fmt.Printf("GitHub Handler Config:\n")
	fmt.Printf("  Client ID: %s\n", maskSensitive(clientID))
	fmt.Printf("  Client Secret: %s\n", maskSensitive(clientSecret))
	fmt.Printf("  Redirect URI: %s\n", redirectURI)

	githubService := services.NewGitHubService(clientID, clientSecret, redirectURI, db)

	// Initialize auth service
	authService, err := auth.NewAuthServiceFromEnv()
	if err != nil {
		fmt.Printf("Warning: Failed to initialize auth service: %v\n", err)
		fmt.Printf("OAuth will not be available\n")
		authService = nil
	}

	return &Handlers{
		tasks:        tasks,
		events:       events,
		agent:        agent,
		github:       githubService,
		auth:         authService,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
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

// RegisterRoutes registers all routes on the router.
func (h *Handlers) RegisterRoutes(r chi.Router) {
	// Landing and Auth
	r.Get("/", h.HandleLanding)
	r.Get("/auth/login", h.HandleAuth)
	r.Get("/auth/register", h.HandleRegister)
	r.Get("/auth/oauth/{provider}", h.HandleOAuth)
	r.Get("/auth/callback", h.HandleAuthCallback)
	r.Get("/auth/check", h.HandleAuthCheck)
	r.Post("/auth/logout", h.HandleLogout)

	// App routes (protected)
	r.Get("/home", h.HandleHome)
	r.Get("/projects", h.HandleProjects)
	r.Post("/projects/refresh", h.HandleRefreshProjects)
	r.Post("/disconnect", h.HandleDisconnect)
	r.Get("/board", h.HandleKanban)
	r.Get("/github/authorize", h.HandleGitHubAuthorize)
	r.Get("/github/callback", h.HandleGitHubCallback)
	r.Get("/tasks", h.HandleListTasks)
	r.Post("/tasks", h.HandleCreateTask)
	r.Post("/tasks/{id}/move", h.HandleMoveTask)
	r.Post("/tasks/{id}/approve", h.HandleApproveTask)
	r.Post("/tasks/{id}/reject", h.HandleRejectTask)
	r.Get("/tasks/{id}/diff", h.HandleTaskDiff)
	r.Get("/tasks/new", h.HandleNewTaskForm)
	r.Get("/events", h.HandleSSE)
}

// HandleHome renders home page.
func (h *Handlers) HandleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if GitHub is connected
	conn, err := h.github.GetActiveConnection(ctx)
	if err == nil && conn != nil {
		// Connected, redirect to projects
		http.Redirect(w, r, "/projects", http.StatusFound)
		return
	}

	// Not connected, show connect page
	component := ui.HomeLayout("counterspell", ui.GitHubConnectPage())
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleProjects renders the projects page.
func (h *Handlers) HandleProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projects, err := h.github.GetProjects(ctx)
	if err != nil {
		projects = []models.Project{}
	}

	// Create a default project if none exist
	if len(projects) == 0 {
		fmt.Println("No projects found, creating default project")
		_ = h.github.SaveProject(ctx, "local", "local-project")
		// Refresh project list
		projects, err = h.github.GetProjects(ctx)
		if err != nil {
			projects = []models.Project{}
		}
	}

	recentProjects, err := h.github.GetRecentProjects(ctx)
	if err != nil {
		recentProjects = []models.Project{}
	}

	conn, err := h.github.GetActiveConnection(ctx)
	if err != nil {
		conn = nil
	}

	component := ui.HomeLayout("counterspell - Projects", ui.ProjectsPage(projects, recentProjects, conn))
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleRefreshProjects fetches repositories from GitHub and returns updated projects page.
func (h *Handlers) HandleRefreshProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get connection
	conn, err := h.github.GetActiveConnection(ctx)
	if err != nil {
		sendHTMXError(w, r, http.StatusNotFound, "No GitHub connection")
		return
	}

	// Fetch and save repositories
	if err := h.github.FetchAndSaveRepositories(ctx, conn); err != nil {
		sendHTMXError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch repositories: %v", err))
		return
	}

	// Get updated projects
	projects, err := h.github.GetProjects(ctx)
	if err != nil {
		projects = []models.Project{}
	}

	recentProjects, err := h.github.GetRecentProjects(ctx)
	if err != nil {
		recentProjects = []models.Project{}
	}

	// Render updated projects section
	w.Header().Set("Content-Type", "text/html")
	if err := ui.ProjectsPage(projects, recentProjects, conn).Render(r.Context(), w); err != nil {
		sendHTMXError(w, r, http.StatusInternalServerError, err.Error())
	}
}

// HandleGitHubAuthorize initiates GitHub OAuth flow.
func (h *Handlers) HandleGitHubAuthorize(w http.ResponseWriter, r *http.Request) {
	connType := r.URL.Query().Get("type") // "org" or "user"

	fmt.Printf("GitHub authorize request - type: %s, clientID: %s\n", connType, h.clientID)

	// Fallback redirect URI if not set
	redirectURI := h.redirectURI
	if redirectURI == "" {
		redirectURI = "http://localhost:8710/github/callback"
	}

	// Redirect to GitHub OAuth
	// URL: https://github.com/login/oauth/authorize
	// Params: client_id, redirect_uri, scope, state
	authURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=repo,read:user,read:org&state=%s",
		h.clientID,
		redirectURI,
		connType)

	fmt.Printf("Redirecting to: %s\n", authURL)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleGitHubCallback handles GitHub OAuth return.
func (h *Handlers) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")
	connType := r.URL.Query().Get("state")

	fmt.Printf("\n=== GitHub OAuth Callback ===\n")
	fmt.Printf("Code: %s\n", maskString(code))
	fmt.Printf("Connection Type: %s\n", connType)

	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	fmt.Printf("Exchanging code for token...\n")
	token, err := h.github.ExchangeCodeForToken(ctx, code)
	if err != nil {
		fmt.Printf("ERROR: Failed to exchange token: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Token received successfully (masked: %s)\n", maskString(token))

	// Get user info
	fmt.Printf("Fetching user info...\n")
	login, avatarURL, err := h.github.GetUserInfo(ctx, token)
	if err != nil {
		fmt.Printf("ERROR: Failed to get user info: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("User info: login=%s, avatar=%s\n", login, avatarURL)

	// Save connection
	fmt.Printf("Saving connection to database...\n")
	err = h.github.SaveConnection(ctx, connType, login, avatarURL, token, "repo,read:user,read:org")
	if err != nil {
		fmt.Printf("ERROR: Failed to save connection: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to save connection: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Connection saved successfully\n")

	// Fetch and save repositories (synchronous for debugging)
	conn, err := h.github.GetActiveConnection(ctx)
	if err == nil && conn != nil {
		fmt.Printf("Fetching repositories for %s (type: %s)\n", conn.Login, conn.Type)
		if err := h.github.FetchAndSaveRepositories(ctx, conn); err != nil {
			fmt.Printf("ERROR: Failed to fetch repositories: %v\n", err)
		} else {
			fmt.Printf("Successfully fetched repositories\n")
		}
	} else {
		fmt.Printf("ERROR: No active connection found: %v\n", err)
	}

	fmt.Printf("=== End OAuth Callback ===\n\n")

	// Redirect to Projects page
	http.Redirect(w, r, "/projects", http.StatusTemporaryRedirect)
}

// maskString masks sensitive strings for logging
func maskString(s string) string {
	if s == "" {
		return "empty"
	}
	if len(s) <= 12 {
		return "***"
	}
	return s[:6] + "..." + s[len(s)-6:]
}

// sendHTMXError sends an error response that triggers HTMX error handling
func sendHTMXError(w http.ResponseWriter, r *http.Request, status int, message string) {
	// Check if this is an HTMX request
	isHTMX := r.Header.Get("HX-Request") == "true"

	if isHTMX {
		// For HTMX requests, return an error response that will trigger the toast
		w.Header().Set("HX-Trigger", "htmx:error")
		http.Error(w, message, status)
	} else {
		// For regular requests, just return the error
		http.Error(w, message, status)
	}
}

// HandleDisconnect disconnects GitHub and clears data.
func (h *Handlers) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("Disconnecting GitHub...\n")

	// Delete connection and projects
	if err := h.github.DeleteConnection(ctx); err != nil {
		fmt.Printf("Failed to delete connection: %v\n", err)
	}
	if err := h.github.DeleteAllProjects(ctx); err != nil {
		fmt.Printf("Failed to delete projects: %v\n", err)
	}

	fmt.Printf("Disconnected successfully, redirecting to home\n")
	// Redirect to home
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// HandleKanban renders the Kanban board.
func (h *Handlers) HandleKanban(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get project ID from URL parameter
	repo := r.URL.Query().Get("repo")

	// Get project for the repo
	projectID := ""
	if repo != "" {
		project, err := h.github.GetProjectByRepo(ctx, repo)
		if err == nil && project != nil {
			projectID = project.ID
		}
	}

	// If no project selected, auto-select the first available project
	if projectID == "" {
		projects, err := h.github.GetProjects(ctx)
		if err == nil && len(projects) > 0 {
			projectID = projects[0].ID
		}
	}

	// Get tasks for all statuses filtered by project
	tasks, err := h.tasks.List(ctx, nil, &projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Group tasks by status
	tasksByStatus := make(map[string][]models.Task)
	for _, task := range tasks {
		tasksByStatus[string(task.Status)] = append(tasksByStatus[string(task.Status)], task)
	}

	// Mock logs (empty for now)
	logsByTask := make(map[string][]models.AgentLog)

	// Determine if mobile
	isMobile := h.isMobile(r)

	component := ui.HomeLayout("counterspell - Kanban", ui.Board(tasks, logsByTask, isMobile, projectID))
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// isMobile checks if request is from mobile device
func (h *Handlers) isMobile(r *http.Request) bool {
	ua := r.UserAgent()
	return strings.Contains(strings.ToLower(ua), "mobile") || 
		strings.Contains(strings.ToLower(ua), "android") ||
		strings.Contains(strings.ToLower(ua), "iphone")
}

// HandleListTasks returns tasks for a specific status column.
func (h *Handlers) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	statusStr := r.URL.Query().Get("status")
	repo := r.URL.Query().Get("repo")

	var status *models.TaskStatus
	if statusStr != "" {
		s := models.TaskStatus(statusStr)
		status = &s
	}

	// Get project ID for the repo
	projectID := ""
	if repo != "" {
		project, err := h.github.GetProjectByRepo(ctx, repo)
		if err == nil && project != nil {
			projectID = project.ID
		}
	}

	tasks, err := h.tasks.List(ctx, status, &projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render task cards as HTML for HTMX
	w.Header().Set("Content-Type", "text/html")
	for _, task := range tasks {
		if _, err := fmt.Fprintf(w, `<div class="task-card bg-white rounded-xl p-4 shadow-sm border border-gray-100 cursor-grab active:cursor-grabbing" data-id="%s">
			<div class="flex items-start justify-between gap-2">
				<h3 class="font-medium text-sm text-gray-900 flex-1">%s</h3>
				<span class="text-xs px-2 py-1 rounded-full bg-gray-100 text-gray-600">%s</span>
			</div>
		</div>`, task.ID, task.Title, task.Status); err != nil {
			// If writing fails, there's nothing we can do
			return
		}
	}
}

// HandleCreateTask creates a new task.
func (h *Handlers) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		sendHTMXError(w, r, http.StatusBadRequest, "Invalid form data: "+err.Error())
		return
	}

	repo := r.FormValue("repo")
	content := r.FormValue("content")

	// Debug logging
	fmt.Printf("HandleCreateTask: repo=%s, content_len=%d\n", repo, len(content))

	if content == "" {
		sendHTMXError(w, r, http.StatusBadRequest, "Content is required")
		return
	}

	// Extract title from markdown content
	title := utils.ExtractTitleFromMarkdown(content)

	// Get project ID for the repo
	projectID := ""
	if repo != "" {
		project, err := h.github.GetProjectByRepo(ctx, repo)
		if err == nil && project != nil {
			projectID = project.ID
		}
	}

	// If still no project ID, try to get first available project
	if projectID == "" {
		projects, err := h.github.GetProjects(ctx)
		if err == nil && len(projects) > 0 {
			projectID = projects[0].ID
		}
	}

	if projectID == "" {
		sendHTMXError(w, r, http.StatusBadRequest, "No project found. Please create a project first.")
		return
	}

	fmt.Printf("Creating task with projectID=%s, title=%s\n", projectID, title)

	task, err := h.tasks.Create(ctx, projectID, title, content)
	if err != nil {
		sendHTMXError(w, r, http.StatusInternalServerError, "Failed to create task: "+err.Error())
		return
	}

	// Return success with task ID to trigger board reload
	w.Header().Set("Content-Type", "text/html")
	component := ui.TaskCreated(task.ID)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleMoveTask updates a task's status.
func (h *Handlers) HandleMoveTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		sendHTMXError(w, r, http.StatusBadRequest, "Task ID is required")
		return
	}

	if err := r.ParseForm(); err != nil {
		sendHTMXError(w, r, http.StatusBadRequest, "Invalid form data")
		return
	}

	statusStr := r.FormValue("status")
	if statusStr == "" {
		sendHTMXError(w, r, http.StatusBadRequest, "Status is required")
		return
	}

	status := models.TaskStatus(statusStr)
	err := h.tasks.UpdateStatus(ctx, taskID, status)
	if err != nil {
		sendHTMXError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// If moving to in_progress, trigger agent
	if status == models.StatusInProgress {
		go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := h.agent.Run(ctx, taskID); err != nil {
			fmt.Printf("Agent run error: %v\n", err)
		}
	}()
	}

	// Return success
	w.WriteHeader(http.StatusOK)
}

// HandleApproveTask approves a task in review and merges changes.
func (h *Handlers) HandleApproveTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Move to done
	err := h.tasks.UpdateStatus(ctx, taskID, models.StatusDone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Merge worktree changes
	// TODO: Clean up worktree

	h.events.Publish(models.Event{
		TaskID:      taskID,
		Type:        "status",
		HTMLPayload: `<span class="text-green-400">✓</span> Task approved and merged`,
	})

	w.WriteHeader(http.StatusOK)
}

// HandleRejectTask rejects a task and rolls back changes.
func (h *Handlers) HandleRejectTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Move back to todo
	err := h.tasks.UpdateStatus(ctx, taskID, models.StatusTodo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Delete worktree

	h.events.Publish(models.Event{
		TaskID:      taskID,
		Type:        "status",
		HTMLPayload: `<span class="text-red-400">✗</span> Task rejected, rolling back`,
	})

	w.WriteHeader(http.StatusOK)
}

// HandleTaskDiff shows the git diff for a task in review.
func (h *Handlers) HandleTaskDiff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	task, err := h.tasks.Get(ctx, taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if task.Status != models.StatusReview {
		http.Error(w, "Task must be in review status", http.StatusBadRequest)
		return
	}

	// Get git diff from worktree
	worktreePath := filepath.Join("..", "worktree-"+taskID)

	cmd := exec.Command("git", "diff", "main")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()

	var diff string
	if err != nil && len(output) > 0 {
		// Could be empty diff (no changes)
		diff = string(output)
	} else if err == nil {
		diff = string(output)
	}

	if diff == "" {
		diff = "No changes to display."
	}

	w.Header().Set("Content-Type", "text/html")
	component := ui.DiffModal(task.Title, diff)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleNewTaskForm renders the new task form modal.
func (h *Handlers) HandleNewTaskForm(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")

	w.Header().Set("Content-Type", "text/html")
	component := ui.NewTaskForm(repo)
	if err := component.Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleSSE streams events for real-time updates.
func (h *Handlers) HandleSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Subscribe to events
	ch := h.events.Subscribe()
	defer h.events.Unsubscribe(ch)

	// Send keepalive every 15s
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				// Connection might be closed, log and return
				return
			}
			flusher.Flush()
		case event := <-ch:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				// Connection might be closed
				return
			}
			flusher.Flush()
		}
	}
}
