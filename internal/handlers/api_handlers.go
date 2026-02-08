package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/revrost/counterspell/internal/models"
	"github.com/revrost/counterspell/internal/services"
)

// HandleListTask returns tasks.
func (h *Handlers) HandleListTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tasks, err := h.repository.ListWithWorkspace(ctx)
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
		case "draft", "in_progress":
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
	taskResp, err := h.repository.GetTaskWithDetails(ctx, taskID)
	if err != nil {
		slog.Error("Failed to get task details", "error", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	render.JSON(w, r, taskResp)
}

// HandleGetTaskDiff returns the git diff for a task.
func (h *Handlers) HandleGetTaskDiff(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}
	gitDiff, err := h.orchestrator.GetDiff(r.Context(), taskID)
	if err != nil {
		slog.Error("Failed to get git diff", "task_id", taskID, "error", err)
		http.Error(w, "Failed to get git diff", http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, map[string]string{"git_diff": gitDiff})
}

// HandleGetSession returns session info based on machine auth status.
func (h *Handlers) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authenticated, identity, err := h.oauthService.IsAuthenticated(ctx)
	if err != nil {
		// Treat transient auth errors as unauthenticated.
		render.JSON(w, r, map[string]any{
			"authenticated":   false,
			"githubConnected": false,
			"needsGitHubAuth": true,
		})
		return
	}
	if !authenticated {
		authErrorCode := h.oauthService.LastLoginErrorCode()
		authErrorMessage := ""
		if authErrorCode == services.LoginErrorOwnerMismatch {
			authErrorMessage = "This Counterspell instance belongs to a different account. Sign in again with the original owner account."
		}

		// No machine JWT found - not authenticated
		render.JSON(w, r, map[string]any{
			"authenticated":    false,
			"githubConnected":  false,
			"needsGitHubAuth":  true,
			"authErrorCode":    authErrorCode,
			"authErrorMessage": authErrorMessage,
		})
		return
	}

	login := ""
	if identity != nil {
		if identity.Subdomain != "" {
			login = identity.Subdomain
		} else if identity.UserID != "" {
			login = identity.UserID
		}
	}

	// User is authenticated via control plane
	render.JSON(w, r, map[string]any{
		"authenticated":   true,
		"githubConnected": true,
		"githubLogin":     login,
		"needsGitHubAuth": false,
	})
}

// HandleFileSearch searches files.
func (h *Handlers) HandleFileSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")
	workspaceID := r.URL.Query().Get("workspace_id")

	if workspaceID == "" {
		slog.Error("Workspace ID required for file search")
		_ = render.Render(w, r, ErrInvalidRequest(errors.New("workspace_id parameter required")))
		return
	}

	workspace, err := h.repository.GetWorkspace(ctx, workspaceID)
	if err != nil {
		slog.Error("Failed to get workspace for file search", "workspace_id", workspaceID, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to get workspace", err))
		return
	}

	filePaths, err := h.fileService.SearchWorkspaceFiles(ctx, workspace.LocalPath, query, 50)
	if err != nil {
		slog.Error("Failed to search workspace files", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to search files", err))
		return
	}

	render.JSON(w, r, filePaths)
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

// HandleNewTask creates a new task from frontend.
func (h *Handlers) HandleNewTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//	userID := "default"

	var req struct {
		Title       string `json:"title"`
		Intent      string `json:"intent"`
		WorkspaceID string `json:"workspace_id"`
		ModelID     string `json:"model_id"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Intent == "" {
		http.Error(w, "Intent required", http.StatusBadRequest)
		return
	}

	fullPrompt := "# " + req.Title + "\n\n" + req.Intent

	slog.Info("[HANDLER] Starting task submission", "workspace_id", req.WorkspaceID, "intent", req.Intent, "model_id", req.ModelID)
	taskID, err := h.orchestrator.NewTask(ctx, req.WorkspaceID, fullPrompt, req.ModelID)
	if err != nil {
		slog.Error("Failed to start task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to start task", err))
		return
	}

	slog.Info("[HANDLER] Task submitted successfully", "task_id", taskID)
	render.JSON(w, r, map[string]string{"task_id": taskID})
}

func (h *Handlers) HandleActionChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Intent  string `json:"intent"`
		TaskID  string `json:"task_id"`
		ModelID string `json:"model_id"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Intent == "" {
		http.Error(w, "Intent required", http.StatusBadRequest)
		return
	}

	slog.Info("[HANDLER] Continue chat submission", "task_id", req.TaskID, "intent", req.Intent, "model_id", req.ModelID)
	err := h.orchestrator.PromptTask(ctx, req.TaskID, req.Intent, req.ModelID)
	if err != nil {
		slog.Error("Failed to start task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to start task", err))
		return
	}

	slog.Info("[HANDLER] Task continued successfully", "task_id", req.TaskID)
	render.JSON(w, r, map[string]string{"task_id": req.TaskID, "status": "in_progress"})
}

// HandleActionClear clears a task.
func (h *Handlers) HandleActionClear(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	if err := h.orchestrator.CleanupTask(r.Context(), taskID); err != nil {
		slog.Error("Failed to clear task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to clear task", err))
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}

// HandleActionMerge attempts to merge task changes.
func (h *Handlers) HandleActionMerge(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	ctx := r.Context()
	//	userID := "default"

	if err := h.orchestrator.MergeTask(ctx, taskID); err != nil {
		slog.Error("Failed to merge task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to merge task", err))
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}

// HandleActionPR creates a pull request for task changes.
func (h *Handlers) HandleActionPR(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	prURL, err := h.orchestrator.CreatePR(r.Context(), taskID)
	if err != nil {
		slog.Error("Failed to create PR", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to create PR", err))
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok", "pr_url": prURL})
}

// HandleActionDiscard discards task changes.
func (h *Handlers) HandleActionDiscard(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	if err := h.orchestrator.CleanupTask(r.Context(), taskID); err != nil {
		slog.Error("Failed to discard task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to discard task", err))
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}
