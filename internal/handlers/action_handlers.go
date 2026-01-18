package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/services"
)

func (h *Handlers) HandleActionRetry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Task restarting...",
	})
}

// HandleActionClear clears a task's chat history and context
func (h *Handlers) HandleActionClear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	if err := h.taskService.ClearHistory(ctx, userID, taskID); err != nil {
		slog.Error("Failed to clear task history", "error", err, "task_id", taskID)
		http.Error(w, "Failed to clear history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Chat history cleared",
	})
}

// HandleActionMerge merges a task's branch to main and pushes
func (h *Handlers) HandleActionMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
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
				http.Error(w, "Failed to load conflict details", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":    "conflict",
				"task_id":   taskID,
				"conflicts": conflicts,
			})
			return
		}

		slog.Error("Failed to merge task", "task_id", taskID, "error", err)
		http.Error(w, "Merge failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Merged to main and pushed!",
	})
}

// HandleResolveConflict resolves a single file conflict
func (h *Handlers) HandleResolveConflict(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	filePath := r.FormValue("file")
	choice := r.FormValue("choice")

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	conflicts, err := orchestrator.GetConflictDetails(ctx, taskID, []string{filePath})
	if err != nil || len(conflicts) == 0 {
		http.Error(w, "Failed to get conflict", http.StatusInternalServerError)
		return
	}

	var resolution string
	switch choice {
	case "ours":
		resolution = conflicts[0].Ours
	case "theirs":
		resolution = conflicts[0].Theirs
	default:
		http.Error(w, "Invalid choice", http.StatusBadRequest)
		return
	}

	if err := orchestrator.ResolveConflict(ctx, taskID, filePath, resolution); err != nil {
		slog.Error("Failed to resolve conflict", "error", err)
		http.Error(w, "Failed to resolve conflict", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Resolved " + filePath,
	})
}

// HandleAbortMerge aborts the current merge
func (h *Handlers) HandleAbortMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := orchestrator.AbortMerge(ctx, taskID); err != nil {
		slog.Error("Failed to abort merge", "error", err)
		http.Error(w, "Failed to abort merge", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Merge aborted",
	})
}

// HandleCompleteMerge completes the merge after conflicts are resolved
func (h *Handlers) HandleCompleteMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := orchestrator.CompleteMergeResolution(ctx, taskID); err != nil {
		slog.Error("Failed to complete merge", "error", err)
		http.Error(w, "Merge failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Merged to main!",
	})
}

// HandleActionPR creates a GitHub Pull Request for the task
func (h *Handlers) HandleActionPR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	prURL, err := orchestrator.CreatePR(ctx, taskID)
	if err != nil {
		slog.Error("Failed to create PR", "task_id", taskID, "error", err)
		http.Error(w, "Failed to create PR: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "PR created!",
		"pr_url":  prURL,
	})
}

// HandleActionDiscard deletes a task and cleans up its worktree
func (h *Handlers) HandleActionDiscard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Clean up worktree first
	if err := orchestrator.CleanupTask(id); err != nil {
		slog.Error("Failed to cleanup worktree", "task_id", id, "error", err)
	}

	if err := h.taskService.Delete(ctx, userID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Task discarded",
	})
}

// HandleActionChat continues a task with a follow-up message.
func (h *Handlers) HandleActionChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(ctx)
	slog.Info("[CHAT] HandleActionChat called", "task_id", taskID)

	if err := r.ParseForm(); err != nil {
		slog.Error("[CHAT] ParseForm failed", "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
	slog.Info("[CHAT] Got message", "message", message)
	if message == "" {
		slog.Warn("[CHAT] Empty message")
		http.Error(w, "Message required", http.StatusBadRequest)
		return
	}

	modelID := r.FormValue("model_id")
	if modelID == "" {
		modelID = "o#anthropic/claude-sonnet-4"
	}

	orchestrator, err := h.getOrchestrator(userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := orchestrator.ContinueTask(ctx, taskID, message, modelID); err != nil {
		slog.Error("Failed to continue task", "task_id", taskID, "error", err)
		http.Error(w, "Failed to continue task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Message sent",
	})
}
