package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// HandleAddTask creates a new task from frontend.
func (h *Handlers) HandleAddTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//	userID := "default"

	var req struct {
		Intent    string `json:"intent"`
		ProjectID string `json:"project_id"`
		ModelID   string `json:"model_id"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Intent == "" {
		http.Error(w, "Intent required", http.StatusBadRequest)
		return
	}

	orch, err := h.getOrchestrator()
	if err != nil {
		slog.Error("Failed to create orchestrator", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to create task", err))
		return
	}

	slog.Info("[HANDLER] Starting task submission", "project_id", req.ProjectID, "intent", req.Intent, "model_id", req.ModelID)
	taskID, err := orch.StartTask(ctx, req.ProjectID, req.Intent, req.ModelID)
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
		Intent   string `json:"intent"`
		TaskID   string `json:"task_id"`
		Model    string `json:"model"`
		Provider string `json:"provider"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Intent == "" {
		http.Error(w, "Intent required", http.StatusBadRequest)
		return
	}

	orch, err := h.getOrchestrator()
	if err != nil {
		slog.Error("Failed to create orchestrator", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to create task", err))
		return
	}

	slog.Info("[HANDLER] Continue chat submission", "task_id", req.TaskID, "intent", req.Intent, "model", req.Model, "provider", req.Provider)
	err = orch.ContinueTask(ctx, req.TaskID, req.Intent, "native", req.Provider, req.Model)
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

	orch, err := h.getOrchestrator()
	if err != nil {
		slog.Error("Failed to create orchestrator", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to clear task", err))
		return
	}

	if err := orch.CleanupTask(taskID); err != nil {
		slog.Error("Failed to clear task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to clear task", err))
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}

// HandleActionRetry retries a failed task.
func (h *Handlers) HandleActionRetry(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	ctx := r.Context()
	//	userID := "default"

	orch, err := h.getOrchestrator()
	if err != nil {
		slog.Error("Failed to create orchestrator", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to retry task", err))
		return
	}

	// For retry, we just start a new task with same intent
	task, err := h.taskService.Get(ctx, taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	repoID := ""
	if task.RepositoryID != nil {
		repoID = *task.RepositoryID
	}
	newTaskID, err := orch.StartTask(ctx, repoID, task.Intent, "")
	if err != nil {
		slog.Error("Failed to retry task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to retry task", err))
		return
	}

	render.JSON(w, r, map[string]string{"task_id": newTaskID})
}

// HandleActionMerge attempts to merge task changes.
func (h *Handlers) HandleActionMerge(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	ctx := r.Context()
	//	userID := "default"

	orch, err := h.getOrchestrator()
	if err != nil {
		slog.Error("Failed to create orchestrator", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to merge task", err))
		return
	}

	if err := orch.MergeTask(ctx, taskID); err != nil {
		slog.Error("Failed to merge task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to merge task", err))
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}

// HandleActionPR creates a pull request for task changes.
func (h *Handlers) HandleActionPR(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	orch, err := h.getOrchestrator()
	if err != nil {
		slog.Error("Failed to create orchestrator", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to create PR", err))
		return
	}

	prURL, err := orch.CreatePR(r.Context(), taskID)
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

	orch, err := h.getOrchestrator()
	if err != nil {
		slog.Error("Failed to create orchestrator", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to discard task", err))
		return
	}

	if err := orch.CleanupTask(taskID); err != nil {
		slog.Error("Failed to discard task", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to discard task", err))
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}
