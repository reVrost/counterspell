package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/revrost/counterspell/internal/models"
)

// HandleSSE handles Server-Sent Events for real-time updates.
func (h *Handlers) HandleSSE(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	ctx := r.Context()
	// Auth removed for local-first mode

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Subscribe to events
	ch := h.events.Subscribe()
	defer h.events.Unsubscribe(ch)

	// Track last sent event ID for client-side deduplication
	var lastSentID int64

	if taskID != "" {
		// Check if task exists
		if _, err := h.taskService.Get(ctx, taskID); err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		// Send initial state
		h.sendInitialState(w, flusher, ctx, taskID)
	} else {
		// Feed page: send initial ping
		_, _ = fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
		flusher.Flush()
	}

	// Keepalive ticker
	keepalive := time.NewTicker(10 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}

			if taskID != "" && event.TaskID != taskID {
				continue
			}

			if event.ID <= lastSentID && event.ID != 0 {
				continue
			}
			lastSentID = event.ID

			// Send event as JSON
			h.sendSSEEvent(w, flusher, event)

		case <-keepalive.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (h *Handlers) sendInitialState(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, taskID string) {
	task, err := h.taskService.Get(ctx, taskID)
	if err != nil {
		return
	}

	// Send full task data as initial state
	data, _ := json.Marshal(task)
	_, _ = fmt.Fprintf(w, "event: initial_state\ndata: %s\n\n", string(data))
	flusher.Flush()
}

func (h *Handlers) sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event models.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal SSE event", "error", err)
		return
	}

	_, _ = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, string(event.Type), string(data))
	flusher.Flush()
}
