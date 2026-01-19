package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/revrost/code/counterspell/internal/services"
)

// HandleAPITasks returns tasks.
func (h *Handlers) HandleAPITasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tasks, err := h.taskService.List(ctx)
	if err != nil {
		slog.Error("Failed to get tasks", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to load tasks", err))
		return
	}
	render.JSON(w, r, tasks)
}

// HandleAPITask returns a single task.
func (h *Handlers) HandleAPITask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	task, err := h.taskService.Get(ctx, taskID)
	if err != nil {
		slog.Error("Failed to get task", "error", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	render.JSON(w, r, task)
}

// HandleAPIMessages returns messages for a task.
func (h *Handlers) HandleAPIMessages(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	messages, err := h.messageService.GetMessagesByTask(ctx, taskID)
	if err != nil {
		slog.Error("Failed to get messages", "error", err)
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, messages)
}

// HandleAPISession returns session info.
func (h *Handlers) HandleAPISession(w http.ResponseWriter, r *http.Request) {
	userID := "default"
	render.JSON(w, r, map[string]any{
		"user_id": userID,
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

// HandleAPISettings returns settings.
func (h *Handlers) HandleAPISettings(w http.ResponseWriter, r *http.Request) {
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
