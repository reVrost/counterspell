package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/revrost/code/counterspell/internal/views"
)

// HandleSSE handles Server-Sent Events for real-time updates.
func (h *Handlers) HandleSSE(w http.ResponseWriter, r *http.Request) {
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

	// Send initial ping
	fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
	flusher.Flush()

	// Stream events
	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}

			data, err := json.Marshal(event)
			if err != nil {
				continue
			}

			fmt.Fprintf(w, "event: task\ndata: %s\n\n", string(data))
			flusher.Flush()
		}
	}
}

// HandleFeedActiveSSE streams active task updates via SSE for htmx.
// This replaces polling with server-push updates.
func (h *Handlers) HandleFeedActiveSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Subscribe to task events
	ch := h.events.Subscribe()
	defer h.events.Unsubscribe(ch)

	// Helper to render and send active tasks
	sendActiveUpdate := func() {
		// Get real projects, fall back to mock if none
		internalProjects, _ := h.github.GetProjects(ctx)
		projects := make(map[string]views.UIProject)
		for _, p := range internalProjects {
			projects[p.ID] = toUIProject(p)
		}

		// Get active tasks from DB
		dbTasks, _ := h.tasks.List(ctx, nil, nil)
		var active []*views.UITask
		for _, t := range dbTasks {
			if t.Status == "in_progress" {
				active = append(active, &views.UITask{
					ID:          t.ID,
					ProjectID:   t.ProjectID,
					Description: t.Title,
					AgentName:   "Agent",
					Status:      string(t.Status),
					Progress:    50, // Stable value to prevent constant morph diffs
				})
			}
		}

		// Add mock data if no real tasks (for demo purposes)
		if len(active) == 0 {
			projects["ios"] = views.UIProject{ID: "ios", Name: "acme/ios-app", Icon: "fa-mobile-alt", Color: "text-green-400"}
			active = append(active, &views.UITask{
				ID:          "3",
				ProjectID:   "ios",
				Description: "Fix crash on startup",
				AgentName:   "Agent-101",
				Status:      "in_progress",
				Progress:    50, // Stable value to prevent constant morph diffs
			})
		}

		// Render to buffer
		var buf bytes.Buffer
		views.ActiveRows(active, projects).Render(ctx, &buf)

		// SSE requires single-line data, escape newlines
		html := strings.ReplaceAll(buf.String(), "\n", "")
		fmt.Fprintf(w, "event: active-update\ndata: %s\n\n", html)
		flusher.Flush()
	}

	// Send initial state
	sendActiveUpdate()

	// Create ticker for periodic updates (fallback)
	// Memory leak was caused by Progress changing every second, not the ticker itself
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Stream updates
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ch:
			if !ok {
				return
			}
			// Task event received, send update
			sendActiveUpdate()
		case <-ticker.C:
			// Periodic refresh as fallback
			sendActiveUpdate()
		}
	}
}
