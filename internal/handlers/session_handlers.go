package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/revrost/counterspell/internal/services"
)

// HandleListSessions returns all sessions.
func (h *Handlers) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessions, err := h.sessionService.List(ctx)
	if err != nil {
		http.Error(w, "Failed to load sessions", http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, sessions)
}

// HandleGetSessionDetail returns a session with messages.
func (h *Handlers) HandleGetSessionDetail(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	session, messages, err := h.sessionService.Get(ctx, sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	render.JSON(w, r, map[string]any{
		"session":  session,
		"messages": messages,
	})
}

// HandleCreateSession creates a new session.
func (h *Handlers) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentBackend string `json:"agent_backend"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	session, err := h.sessionService.Create(ctx, req.AgentBackend)
	if err != nil {
		if errors.Is(err, services.ErrCodexUnsupported) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, session)
}

// HandleSessionChat sends a chat message to a session.
func (h *Handlers) HandleSessionChat(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	var req struct {
		Message string `json:"message"`
		ModelID string `json:"model_id"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		http.Error(w, "Message required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.sessionService.Chat(ctx, sessionID, req.Message, req.ModelID); err != nil {
		if errors.Is(err, services.ErrCodexUnsupported) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}

// HandlePromoteSession promotes a session to a task.
func (h *Handlers) HandlePromoteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	task, err := h.sessionService.Promote(ctx, sessionID)
	if err != nil {
		if errors.Is(err, services.ErrCodexUnsupported) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to promote session", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]string{"task_id": task.ID})
}
