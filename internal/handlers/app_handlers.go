// Package handlers contains the handlers for the API.
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/models"
)

// HandleHome returns a JSON object with app metadata
func (h *Handlers) HandleHome(w http.ResponseWriter, r *http.Request) {
	response := struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
	}{
		Name:        "Counterspell",
		Version:     "2.1.0",
		Description: "Mobile-first, hosted AI agent Kanban.",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// HandleTaskDetailUI returns task details as JSON (alias to HandleAPITask)
func (h *Handlers) HandleTaskDetailUI(w http.ResponseWriter, r *http.Request) {
	h.HandleAPITask(w, r)
}

// HandleActionRetry mocks retry action

// HandleAddTask creates a new task and starts execution.
func (h *Handlers) HandleAddTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	intent := r.FormValue("voice_input")
	projectID := r.FormValue("project_id")
	modelID := r.FormValue("model_id")

	if intent == "" {
		// h.HandleFeed(w, r)
		return
	}

	if projectID == "" {
		w.Header().Set("HX-Trigger", `{"toast": "Select a project first", "taskCreated": "false"}`)
		return
	}

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		slog.Error("Failed to get orchestrator", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Internal error", "taskCreated": "false"}`)
		return
	}

	_, err = orchestrator.StartTask(ctx, projectID, intent, modelID)
	if err != nil {
		slog.Error("Failed to start task", "error", err)
		http.Error(w, "Failed to start task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "message": "Task started"})
}

// HandleSaveSettings saves user settings
func (h *Handlers) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if userID == "" {
		userID = "default"
	}

	settings := &models.UserSettings{
		UserID:        userID,
		OpenRouterKey: r.FormValue("openrouter_key"),
		ZaiKey:        r.FormValue("zai_key"),
		AnthropicKey:  r.FormValue("anthropic_key"),
		OpenAIKey:     r.FormValue("openai_key"),
		AgentBackend:  r.FormValue("agent_backend"),
	}

	if err := h.settingsService.UpdateSettings(ctx, userID, settings); err != nil {
		http.Error(w, "Failed to save settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "message": "Settings saved successfully"})
}

// HandleFileSearch searches for files in a project using fuzzy matching.
func (h *Handlers) HandleFileSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)
	projectID := r.URL.Query().Get("project_id")
	query := r.URL.Query().Get("q")

	if projectID == "" {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
		return
	}

	files, err := orchestrator.SearchProjectFiles(ctx, projectID, query, 20)
	if err != nil {
		slog.Error("File search failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
		return
	}

	if files == nil {
		files = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		slog.Error("Failed to encode file search results", "error", err)
	}
}

// HandleTranscribe transcribes uploaded audio to text.
func (h *Handlers) HandleTranscribe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "No audio file provided", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")

	text, err := h.transcription.TranscribeAudio(ctx, file, contentType)
	if err != nil {
		slog.Error("Transcription failed", "error", err)
		http.Error(w, "Transcription failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//nolint:errcheck
	w.Write([]byte(text))
}
