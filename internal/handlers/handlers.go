package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/services"
	"github.com/revrost/code/counterspell/internal/ui"
)

// Handlers contains all HTTP handlers.
type Handlers struct {
	tasks       *services.TaskService
	events      *services.EventBus
	agent       *services.AgentRunner
	clientID    string
	clientSecret string
	redirectURI  string
}

// NewHandlers creates new HTTP handlers.
func NewHandlers(tasks *services.TaskService, events *services.EventBus, agent *services.AgentRunner) *Handlers {
	return &Handlers{
		tasks:       tasks,
		events:      events,
		agent:       agent,
		clientID:    os.Getenv("GITHUB_CLIENT_ID"),
		clientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		redirectURI:  os.Getenv("GITHUB_REDIRECT_URI"),
	}
}

// RegisterRoutes registers all routes on the router.
func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Get("/", h.HandleHome)
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
	component := ui.HomeLayout("counterspell", ui.GitHubConnectPage())
	component.Render(r.Context(), w)
}

// HandleGitHubAuthorize initiates GitHub OAuth flow.
func (h *Handlers) HandleGitHubAuthorize(w http.ResponseWriter, r *http.Request) {
	connType := r.URL.Query().Get("type") // "org" or "user"
	
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
	
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleGitHubCallback handles GitHub OAuth return.
func (h *Handlers) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	// Log the auth code (in production, remove this!)
	fmt.Printf("GitHub OAuth Code: %s, State: %s\n", code, state)

	// Exchange code for access_token
	// POST https://github.com/login/oauth/access_token
	// Headers: Accept: application/json
	// Body: client_id, client_secret, code, redirect_uri
	//
	// Response:
	// {
	//   "access_token": "gho_...",
	//   "token_type": "bearer",
	//   "scope": "repo,read:user,read:org"
	// }
	//
	// TODO: Implement actual token exchange
	// For now, we'll skip the exchange and just redirect

	// Redirect to Kanban board
	http.Redirect(w, r, "/board", http.StatusTemporaryRedirect)
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
	component.Render(r.Context(), w)
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
		fmt.Fprintf(w, `<div class="task-card bg-white rounded-xl p-4 shadow-sm border border-gray-100 cursor-grab active:cursor-grabbing" data-id="%s">
			<div class="flex items-start justify-between gap-2">
				<h3 class="font-medium text-sm text-gray-900 flex-1">%s</h3>
				<span class="text-xs px-2 py-1 rounded-full bg-gray-100 text-gray-600">%s</span>
			</div>
		</div>`, task.ID, task.Title, task.Status)
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
	fmt.Fprintf(w, `<div class="task-card bg-white rounded-xl p-4 shadow-sm border border-gray-100" data-id="%s">
		<h3 class="font-medium text-sm text-gray-900">%s</h3>
	</div>`, task.ID, task.Title)
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
		go h.agent.Run(ctx, taskID)
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
	fmt.Fprint(w, `<div class="bg-white rounded-xl p-6 max-w-md w-full shadow-xl" hx-post="/tasks" hx-target="#main" hx-swap="outerHTML">
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
	</div>`)
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
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case event := <-ch:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
