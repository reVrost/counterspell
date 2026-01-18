// Package handlers contains the handlers for the API.
package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/models"
)

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
		_ = render.Render(w, r, ErrInternalServer("Failed to get orchestrator"))
		return
	}

	_, err = orchestrator.StartTask(ctx, data.ProjectID, data.VoiceInput, data.ModelID)
	if err != nil {
		slog.Error("Failed to start task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to start task: "+err.Error()))
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
		_ = render.Render(w, r, ErrInternalServer("Failed to save settings: "+err.Error()))
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
		_ = render.Render(w, r, ErrInternalServer("Transcription failed: "+err.Error()))
		return
	}

	render.PlainText(w, r, text)
}
