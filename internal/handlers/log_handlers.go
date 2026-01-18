package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/render"
	"github.com/revrost/code/counterspell/internal/auth"
)

// UILogEntry represents a log entry from the UI.
type UILogEntry struct {
	Level     string         `json:"level"`     // error, warn, info, debug
	Message   string         `json:"message"`   // Main message
	Component string         `json:"component"` // Svelte component name
	Stack     string         `json:"stack"`     // Stack trace if available
	URL       string         `json:"url"`       // Current page URL
	Extra     map[string]any `json:"extra"`     // Additional context
}

func (u *UILogEntry) Bind(r *http.Request) error {
	return nil
}

// HandleUILog receives log entries from the UI and writes them to server log via slog.
func (h *Handlers) HandleUILog(w http.ResponseWriter, r *http.Request) {
	// userID may be empty if auth failed - that's OK for logging
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		userID = "anonymous"
	}

	entry := &UILogEntry{}
	if err := render.Bind(r, entry); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Build slog attributes
	attrs := []any{
		"origin", "ui",
		"user_id", userID,
		"component", entry.Component,
		"url", entry.URL,
	}
	if entry.Stack != "" {
		attrs = append(attrs, "stack", entry.Stack)
	}
	for k, v := range entry.Extra {
		attrs = append(attrs, k, v)
	}

	// Log at appropriate level
	switch entry.Level {
	case "error":
		slog.Error(entry.Message, attrs...)
	case "warn":
		slog.Warn(entry.Message, attrs...)
	case "debug":
		slog.Debug(entry.Message, attrs...)
	default:
		slog.Info(entry.Message, attrs...)
	}

	render.NoContent(w, r)
}

// HandleReadLogs returns server.log contents for agent debugging.
func (h *Handlers) HandleReadLogs(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("server.log")
	if err != nil {
		_ = render.Render(w, r, ErrNotFound("Log file not found"))
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/plain")
	_, _ = io.Copy(w, f)
}
