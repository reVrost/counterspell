package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

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

// HandleUILog receives log entries from the UI and writes them to server log via slog.
func (h *Handlers) HandleUILog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// userID may be empty if auth failed - that's OK for logging
	userID := auth.UserIDFromContext(r.Context())
	if userID == "" {
		userID = "anonymous"
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var entry UILogEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Build slog attributes
	attrs := []any{
		"source", "ui",
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

	w.WriteHeader(http.StatusNoContent)
}

// HandleReadLogs returns server.log contents for agent debugging.
func (h *Handlers) HandleReadLogs(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("server.log")
	if err != nil {
		http.Error(w, "Log file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(data)
}
