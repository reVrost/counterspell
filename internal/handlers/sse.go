package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/models"
)

// HandleSSE handles Server-Sent Events for real-time updates.
// Supports filtering by task_id via query parameter.
// If task_id is provided, only sends events for that task.
// Otherwise, sends all events (used by feed page).
func (h *Handlers) HandleSSE(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)

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
		if _, err := h.taskService.Get(ctx, userID, taskID); err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		// Send initial state
		h.sendInitialState(w, flusher, ctx, userID, taskID)
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

func (h *Handlers) sendInitialState(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, userID, taskID string) {
	task, err := h.taskService.Get(ctx, userID, taskID)
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

	// For Svelte/Frontend consumption, we can use the Event ID and Type as SSE fields
	// but the payload should be the full event object for consistency
	_, _ = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, string(event.Type), string(data))
	flusher.Flush()
}

// renderDiffHTML converts git diff text to GitHub-styled HTML
// func renderDiffHTML(diff string) string {
// 	if diff == "" {
// 		return `<div class="text-gray-500 italic">No changes made</div>`
// 	}
//
// 	var buf strings.Builder
// 	var currentFile string
// 	var lineNum int
// 	inFileBlock := false
//
// 	for line := range strings.SplitSeq(diff, "\n") {
// 		escapedLine := escapeHTML(line)
//
// 		if strings.HasPrefix(line, "diff --git") {
// 			// Close previous file block
// 			if inFileBlock {
// 				buf.WriteString(`</div>`)
// 			}
// 			inFileBlock = true
// 			lineNum = 0
//
// 			// Extract filename
// 			parts := strings.Split(line, " b/")
// 			if len(parts) > 1 {
// 				currentFile = parts[len(parts)-1]
// 			}
// 			fmt.Fprintf(&buf, `<div class="diff-file-header"><i class="fas fa-file-code mr-2 text-gray-500"></i>%s</div><div class="diff-file-body">`, escapeHTML(currentFile))
// 		} else if strings.HasPrefix(line, "@@") {
// 			// Hunk header
// 			fmt.Fprintf(&buf, `<div class="diff-hunk">%s</div>`, escapedLine)
// 		} else if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "index ") {
// 			// Skip meta lines
// 		} else if strings.HasPrefix(line, "+") {
// 			lineNum++
// 			fmt.Fprintf(&buf, `<div class="diff-line diff-add"><span class="diff-line-num">%d</span><span class="diff-line-content">%s</span></div>`, lineNum, escapedLine)
// 		} else if strings.HasPrefix(line, "-") {
// 			fmt.Fprintf(&buf, `<div class="diff-line diff-del"><span class="diff-line-num"></span><span class="diff-line-content">%s</span></div>`, escapedLine)
// 		} else if line != "" {
// 			lineNum++
// 			fmt.Fprintf(&buf, `<div class="diff-line diff-context"><span class="diff-line-num">%d</span><span class="diff-line-content">%s</span></div>`, lineNum, escapedLine)
// 		}
// 	}
//
// 	if inFileBlock {
// 		buf.WriteString(`</div>`)
// 	}
//
// 	return buf.String()
// }

// escapeHTML escapes HTML special characters
// func escapeHTML(s string) string {
// 	s = strings.ReplaceAll(s, "&", "&amp;")
// 	s = strings.ReplaceAll(s, "<", "&lt;")
// 	s = strings.ReplaceAll(s, ">", "&gt;")
// 	s = strings.ReplaceAll(s, "\"", "&quot;")
// 	s = strings.ReplaceAll(s, "'", "&#39;")
// 	return s
// }
