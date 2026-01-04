package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/services"
	"github.com/revrost/code/counterspell/internal/ui"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	tasks        *services.TaskService
	events       *services.EventBus
	agent        *services.AgentRunner
	github       *services.GitHubService
	clientID     string
	clientSecret string
	redirectURI  string
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(tasks *services.TaskService, events *services.EventBus, agent *services.AgentRunner, db *db.DB) *Handlers {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("GITHUB_REDIRECT_URI")

	return &Handlers{
		tasks:        tasks,
		events:       events,
		agent:        agent,
		github:       services.NewGitHubService(clientID, clientSecret, redirectURI, db),
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

// RegisterRoutes registers all routes on the router.
func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Get("/", h.HandleHome)
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
		http.Error(w, "No GitHub connection", http.StatusNotFound)
		return
	}

	// Fetch and save repositories
	if err := h.github.FetchAndSaveRepositories(ctx, conn); err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch repositories: %v", err), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	token, err := h.github.ExchangeCodeForToken(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Token received successfully\n")

	// Get user info
	login, avatarURL, err := h.github.GetUserInfo(ctx, token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("User info: login=%s\n", login)

	// Save connection
	err = h.github.SaveConnection(ctx, connType, login, avatarURL, token, "repo,read:user,read:org")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save connection: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Connection saved\n")

	// Fetch and save repositories (synchronous for debugging)
	conn, err := h.github.GetActiveConnection(ctx)
	if err == nil && conn != nil {
		fmt.Printf("Fetching repositories for %s (type: %s)\n", conn.Login, conn.Type)
		if err := h.github.FetchAndSaveRepositories(ctx, conn); err != nil {
			fmt.Printf("Failed to fetch repositories: %v\n", err)
		} else {
			fmt.Printf("Successfully fetched repositories\n")
		}
	} else {
		fmt.Printf("No active connection found: %v\n", err)
	}

	// Redirect to Projects page
	http.Redirect(w, r, "/projects", http.StatusTemporaryRedirect)
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
	
	// Get tasks for all statuses
	tasks, err := h.tasks.List(ctx, nil)
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
	
	component := ui.HomeLayout("counterspell - Kanban", ui.Board(tasks, logsByTask, isMobile))
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

	var status *models.TaskStatus
	if statusStr != "" {
		s := models.TaskStatus(statusStr)
		status = &s
	}

	tasks, err := h.tasks.List(ctx, status)
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
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	intent := r.FormValue("intent")

	if title == "" || intent == "" {
		http.Error(w, "Title and intent are required", http.StatusBadRequest)
		return
	}

	task, err := h.tasks.Create(ctx, title, intent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the new task as HTML
	w.Header().Set("Content-Type", "text/html")
	if _, err := fmt.Fprintf(w, `<div class="task-card bg-white rounded-xl p-4 shadow-sm border border-gray-100" data-id="%s">
		<h3 class="font-medium text-sm text-gray-900">%s</h3>
	</div>`, task.ID, task.Title); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleMoveTask updates a task's status.
func (h *Handlers) HandleMoveTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	statusStr := r.FormValue("status")
	if statusStr == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	status := models.TaskStatus(statusStr)
	err := h.tasks.UpdateStatus(ctx, taskID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

// HandleNewTaskForm renders the new task form modal.
func (h *Handlers) HandleNewTaskForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	form := `<div class="bg-white rounded-xl p-6 max-w-md w-full shadow-xl" hx-post="/tasks" hx-target="#main" hx-swap="outerHTML">
		<h2 class="text-lg font-semibold mb-4 tracking-tighter">New Task</h2>
		<form class="space-y-4">
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
				<input type="text" name="title" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-black" placeholder="Fix login bug"/>
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700 mb-1">Intent</label>
				<textarea name="intent" class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-black" rows="3" placeholder="User cannot login with email..."></textarea>
			</div>
			<div class="flex gap-2">
				<button type="submit" class="flex-1 bg-black text-white py-2 rounded-lg font-medium active:scale-95 transition-transform">Create Task</button>
				<button type="button" hx-target="#modal" hx-swap="innerHTML" class="flex-1 bg-gray-100 text-gray-900 py-2 rounded-lg font-medium">Cancel</button>
			</div>
		</form>
	</div>`
	if _, err := fmt.Fprint(w, form); err != nil {
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
