package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/services"
)

// HandleListTask returns tasks.
func (h *Handlers) HandleListTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tasks, err := h.taskService.ListWithRepository(ctx)
	if err != nil {
		slog.Error("Failed to get tasks", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to load tasks", err))
		return
	}

	feed := &FeedData{
		Active:   []*models.Task{},
		Reviews:  []*models.Task{},
		Done:     []*models.Task{},
		Todo:     []*models.Task{},
		Planning: []*models.Task{},
	}

	for _, t := range tasks {
		switch t.Status {
		case "pending", "in_progress":
			feed.Active = append(feed.Active, t)
		case "planning":
			feed.Planning = append(feed.Planning, t)
		case "review":
			feed.Reviews = append(feed.Reviews, t)
		case "done", "failed":
			feed.Done = append(feed.Done, t)
		}
	}

	if err := render.Render(w, r, feed); err != nil {
		http.Error(w, "Failed to render response", http.StatusInternalServerError)
		return
	}
}

// HandleGetTask returns a single task with full details including messages and artifacts.
func (h *Handlers) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	taskResp, err := h.taskService.GetTaskWithDetails(ctx, taskID)
	if err != nil {
		slog.Error("Failed to get task details", "error", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Get git diff if task is in progress or review
	if taskResp.Task.Status == "in_progress" || taskResp.Task.Status == "review" {
		gitDiff, err := h.gitReposManager.GetDiff(taskID)
		if err != nil {
			slog.Warn("Failed to get git diff", "task_id", taskID, "error", err)
			// Continue without diff, don't fail the request
		} else {
			taskResp.GitDiff = gitDiff
		}
	}

	render.JSON(w, r, taskResp)
}

// HandleGetSession returns session info based on GitHub connection status.
func (h *Handlers) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if there's a GitHub connection
	conn, err := h.githubService.GetConnection(ctx)
	if err != nil {
		// No connection found - not authenticated
		render.JSON(w, r, map[string]any{
			"authenticated":   false,
			"githubConnected": false,
			"needsGitHubAuth": true,
		})
		return
	}

	// User is authenticated via GitHub
	render.JSON(w, r, map[string]any{
		"authenticated":   true,
		"githubConnected": true,
		"githubLogin":     conn.Username,
		"needsGitHubAuth": false,
	})
}

// HandleFileSearch searches files.
func (h *Handlers) HandleFileSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	directory := r.URL.Query().Get("directory")

	ctx := r.Context()
	files, err := h.fileService.Search(ctx, query, directory, 50)
	if err != nil {
		slog.Error("Failed to search files", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to search files", err))
		return
	}

	render.JSON(w, r, files)
}

// HandleGetSettings returns settings.
func (h *Handlers) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	settings, err := h.settingsService.GetSettings(ctx)
	if err != nil {
		slog.Error("Failed to get settings", "error", err)
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, settings)
}

// HandleSaveSettings saves settings with validation.
func (h *Handlers) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	var settings services.Settings
	if err := render.DecodeJSON(r.Body, &settings); err != nil {
		slog.Error("Failed to decode settings", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.settingsService.UpdateSettings(ctx, &settings); err != nil {
		slog.Error("Failed to save settings", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}

// HandleTranscribe handles transcription.
func (h *Handlers) HandleTranscribe(w http.ResponseWriter, r *http.Request) {
	// Placeholder
	render.JSON(w, r, map[string]string{"status": "not implemented"})
}
