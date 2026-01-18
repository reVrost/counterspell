package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/services"
)

func (h *Handlers) HandleActionRetry(w http.ResponseWriter, r *http.Request) {
	_ = render.Render(w, r, Success("Task restarting..."))
}

// HandleActionClear clears a task's chat history and context
func (h *Handlers) HandleActionClear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	if err := h.taskService.ClearHistory(ctx, userID, taskID); err != nil {
		slog.Error("Failed to clear task history", "error", err, "task_id", taskID)
		_ = render.Render(w, r, ErrInternalServer("Failed to clear history"))
		return
	}

	_ = render.Render(w, r, Success("Chat history cleared"))
}

// HandleActionMerge merges a task's branch to main and pushes
func (h *Handlers) HandleActionMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Internal error"))
		return
	}

	err = orchestrator.MergeTask(ctx, taskID)
	if err != nil {
		// Check if it's a merge conflict
		if conflictErr, ok := err.(*services.ErrMergeConflict); ok {
			slog.Info("Merge conflict detected", "task_id", taskID, "files", conflictErr.ConflictedFiles)

			conflicts, err := orchestrator.GetConflictDetails(ctx, taskID, conflictErr.ConflictedFiles)
			if err != nil {
				slog.Error("Failed to get conflict details", "error", err)
				_ = render.Render(w, r, ErrInternalServer("Failed to load conflict details"))
				return
			}

			_ = render.Render(w, r, Conflict(taskID, conflicts))
			return
		}

		slog.Error("Failed to merge task", "task_id", taskID, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Merge failed: "+err.Error()))
		return
	}

	_ = render.Render(w, r, Success("Merged to main and pushed!"))
}

// HandleResolveConflict resolves a single file conflict
func (h *Handlers) HandleResolveConflict(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	data := &ResolveConflictRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Internal error"))
		return
	}

	conflicts, err := orchestrator.GetConflictDetails(ctx, taskID, []string{data.File})
	if err != nil || len(conflicts) == 0 {
		_ = render.Render(w, r, ErrInternalServer("Failed to get conflict"))
		return
	}

	var resolution string
	switch data.Choice {
	case "ours":
		resolution = conflicts[0].Ours
	case "theirs":
		resolution = conflicts[0].Theirs
	}

	if err := orchestrator.ResolveConflict(ctx, taskID, data.File, resolution); err != nil {
		slog.Error("Failed to resolve conflict", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to resolve conflict"))
		return
	}

	_ = render.Render(w, r, Success("Resolved "+data.File))
}

// HandleAbortMerge aborts the current merge
func (h *Handlers) HandleAbortMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Internal error"))
		return
	}

	if err := orchestrator.AbortMerge(ctx, taskID); err != nil {
		slog.Error("Failed to abort merge", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to abort merge"))
		return
	}

	_ = render.Render(w, r, Success("Merge aborted"))
}

// HandleCompleteMerge completes the merge after conflicts are resolved
func (h *Handlers) HandleCompleteMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Internal error"))
		return
	}

	if err := orchestrator.CompleteMergeResolution(ctx, taskID); err != nil {
		slog.Error("Failed to complete merge", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Merge failed: "+err.Error()))
		return
	}

	_ = render.Render(w, r, Success("Merged to main!"))
}

// HandleActionPR creates a GitHub Pull Request for the task
func (h *Handlers) HandleActionPR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Internal error"))
		return
	}

	prURL, err := orchestrator.CreatePR(ctx, taskID)
	if err != nil {
		slog.Error("Failed to create PR", "task_id", taskID, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to create PR: "+err.Error()))
		return
	}

	_ = render.Render(w, r, SuccessWithPR("PR created!", prURL))
}

// HandleActionDiscard deletes a task and cleans up its worktree
func (h *Handlers) HandleActionDiscard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Internal error"))
		return
	}

	// Clean up worktree first
	if err := orchestrator.CleanupTask(id); err != nil {
		slog.Error("Failed to cleanup worktree", "task_id", id, "error", err)
	}

	if err := h.taskService.Delete(ctx, userID, id); err != nil {
		_ = render.Render(w, r, ErrInternalServer(err.Error()))
		return
	}

	_ = render.Render(w, r, Success("Task discarded"))
}

// HandleActionChat continues a task with a follow-up message.
func (h *Handlers) HandleActionChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)
	slog.Info("[CHAT] HandleActionChat called", "task_id", taskID)

	data := &ChatRequest{}
	if err := render.Bind(r, data); err != nil {
		slog.Error("[CHAT] Bind failed", "error", err)
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	slog.Info("[CHAT] Got message", "message", data.Message)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Internal error"))
		return
	}

	if err := orchestrator.ContinueTask(ctx, taskID, data.Message, data.ModelID); err != nil {
		slog.Error("Failed to continue task", "task_id", taskID, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to continue task"))
		return
	}

	_ = render.Render(w, r, Success("Message sent"))
}
