package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/services"
	"github.com/revrost/code/counterspell/internal/views/components"
)

func (h *Handlers) HandleActionRetry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Task restarting..."}`)
	w.WriteHeader(http.StatusNoContent)
}

// HandleActionClear clears a task's chat history and context
func (h *Handlers) HandleActionClear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	svc, err := h.getServices(ctx)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := svc.Tasks.ClearHistory(ctx, taskID); err != nil {
		slog.Error("Failed to clear task history", "error", err, "task_id", taskID)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to clear history"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Chat history cleared"}`)
	w.WriteHeader(http.StatusNoContent)
}

// HandleActionMerge merges a task's branch to main and pushes
func (h *Handlers) HandleActionMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	orchestrator, err := h.getOrchestrator(ctx)
	if err != nil {
		slog.Error("Failed to get orchestrator", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = orchestrator.MergeTask(ctx, taskID)
	if err != nil {
		// Check if it's a merge conflict
		if conflictErr, ok := err.(*services.ErrMergeConflict); ok {
			slog.Info("Merge conflict detected, showing conflict UI", "task_id", taskID, "files", conflictErr.ConflictedFiles)

			conflicts, err := orchestrator.GetConflictDetails(ctx, taskID, conflictErr.ConflictedFiles)
			if err != nil {
				slog.Error("Failed to get conflict details", "error", err)
				w.Header().Set("HX-Trigger", `{"toast": "Failed to load conflict details"}`)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			uiConflicts := make([]components.ConflictFile, len(conflicts))
			for i, c := range conflicts {
				uiConflicts[i] = components.ConflictFile{
					Path:   c.Path,
					Ours:   c.Ours,
					Theirs: c.Theirs,
				}
			}

			if err := components.ConflictView(taskID, uiConflicts).Render(ctx, w); err != nil {
				slog.Error("render error", "error", err)
			}
			return
		}

		slog.Error("Failed to merge task", "task_id", taskID, "error", err)
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"toast": "Merge failed: %s"}`, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Merged to main and pushed!"}`)
	w.WriteHeader(http.StatusNoContent)
}

// HandleResolveConflict resolves a single file conflict
func (h *Handlers) HandleResolveConflict(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Invalid request"}`)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filePath := r.FormValue("file")
	choice := r.FormValue("choice")

	orchestrator, err := h.getOrchestrator(ctx)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	conflicts, err := orchestrator.GetConflictDetails(ctx, taskID, []string{filePath})
	if err != nil || len(conflicts) == 0 {
		w.Header().Set("HX-Trigger", `{"toast": "Failed to get conflict"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var resolution string
	switch choice {
	case "ours":
		resolution = conflicts[0].Ours
	case "theirs":
		resolution = conflicts[0].Theirs
	default:
		w.Header().Set("HX-Trigger", `{"toast": "Invalid choice"}`)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := orchestrator.ResolveConflict(ctx, taskID, filePath, resolution); err != nil {
		slog.Error("Failed to resolve conflict", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to resolve conflict"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"toast": "Resolved %s"}`, filePath))
	w.WriteHeader(http.StatusOK)
}

// HandleAbortMerge aborts the current merge
func (h *Handlers) HandleAbortMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	orchestrator, err := h.getOrchestrator(ctx)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := orchestrator.AbortMerge(ctx, taskID); err != nil {
		slog.Error("Failed to abort merge", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to abort merge"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Merge aborted"}`)
	h.RenderApp(w, r)
}

// HandleCompleteMerge completes the merge after conflicts are resolved
func (h *Handlers) HandleCompleteMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	orchestrator, err := h.getOrchestrator(ctx)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := orchestrator.CompleteMergeResolution(ctx, taskID); err != nil {
		slog.Error("Failed to complete merge", "error", err)
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"toast": "Merge failed: %s"}`, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Merged to main!"}`)
	h.RenderApp(w, r)
}

// HandleActionPR creates a GitHub Pull Request for the task
func (h *Handlers) HandleActionPR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	orchestrator, err := h.getOrchestrator(ctx)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	prURL, err := orchestrator.CreatePR(ctx, taskID)
	if err != nil {
		slog.Error("Failed to create PR", "task_id", taskID, "error", err)
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"toast": "Failed to create PR: %s"}`, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"close-modal": true, "toast": "PR created!", "open-url": "%s"}`, prURL))
	w.WriteHeader(http.StatusNoContent)
}

// HandleActionDiscard deletes a task and cleans up its worktree
func (h *Handlers) HandleActionDiscard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	svc, err := h.getServices(ctx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	orchestrator, err := h.getOrchestrator(ctx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Clean up worktree first
	if err := orchestrator.CleanupTask(id); err != nil {
		slog.Error("Failed to cleanup worktree", "task_id", id, "error", err)
	}

	if err := svc.Tasks.Delete(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send trigger only - the close-modal event will cause the feed to refresh via hx-trigger
	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Task discarded"}`)
	w.WriteHeader(http.StatusNoContent)
}

// HandleActionChat continues a task with a follow-up message.
func (h *Handlers) HandleActionChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	slog.Info("[CHAT] HandleActionChat called", "task_id", taskID)

	if err := r.ParseForm(); err != nil {
		slog.Error("[CHAT] ParseForm failed", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Invalid request"}`)
		h.RenderApp(w, r)
		return
	}

	message := r.FormValue("message")
	slog.Info("[CHAT] Got message", "message", message)
	if message == "" {
		slog.Warn("[CHAT] Empty message")
		w.Header().Set("HX-Trigger", `{"toast": "Message required"}`)
		h.RenderApp(w, r)
		return
	}

	modelID := r.FormValue("model_id")
	if modelID == "" {
		modelID = "o#anthropic/claude-sonnet-4"
	}

	orchestrator, err := h.getOrchestrator(ctx)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		h.RenderApp(w, r)
		return
	}

	if err := orchestrator.ContinueTask(ctx, taskID, message, modelID); err != nil {
		slog.Error("Failed to continue task", "task_id", taskID, "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to continue task"}`)
		h.RenderApp(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}
