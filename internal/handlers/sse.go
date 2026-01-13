package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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

	// Helper to render and send active tasks and reviews
	sendActiveUpdate := func() {
		// Get real projects
		internalProjects, _ := h.github.GetProjects(ctx)
		projects := make(map[string]views.UIProject)
		for _, p := range internalProjects {
			projects[p.ID] = toUIProject(p)
		}

		// Get tasks from DB and categorize
		dbTasks, _ := h.tasks.List(ctx, nil, nil)
		var active []*views.UITask
		var reviews []*views.UITask
		for _, t := range dbTasks {
			uiTask := &views.UITask{
				ID:          t.ID,
				ProjectID:   t.ProjectID,
				Description: t.Title,
				AgentName:   "Agent",
				Status:      string(t.Status),
				Progress:    50,
			}
			if t.Status == "in_progress" {
				active = append(active, uiTask)
			} else if t.Status == "review" || t.Status == "human_review" {
				uiTask.Progress = 100
				reviews = append(reviews, uiTask)
			}
		}

		// Render active rows to buffer
		var buf bytes.Buffer
		views.ActiveRows(active, projects).Render(ctx, &buf)

		// Add OOB swap for reviews section
		buf.WriteString(`<div id="reviews-container" hx-swap-oob="true">`)
		views.ReviewsSection(views.FeedData{Reviews: reviews, Projects: projects}).Render(ctx, &buf)
		buf.WriteString(`</div>`)

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

// HandleTaskLogsSSE streams log updates for a specific task.
func (h *Handlers) HandleTaskLogsSSE(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

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

	// Subscribe to events
	ch := h.events.Subscribe()
	defer h.events.Unsubscribe(ch)

	// Send initial ping
	fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
	flusher.Flush()

	// Stream only logs for this task
	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}

			// Only forward events for this task
			if event.TaskID != taskID {
				continue
			}

			// Send log event as HTML for htmx SSE
			if event.Type == "log" {
				// The HTMLPayload already has formatted content like: <span class="text-yellow-400">[plan]</span> message
				// Wrap it in log entry structure and send
				htmlData := strings.ReplaceAll(event.HTMLPayload, "\n", "")
				logHTML := fmt.Sprintf(`<div class="ml-4 relative"><div class="absolute -left-[21px] top-1 h-2.5 w-2.5 rounded-full border border-[#0D1117] bg-blue-500"></div><p class="text-xs text-gray-400">%s</p></div>`, htmlData)
				fmt.Fprintf(w, "event: log\ndata: %s\n\n", logHTML)
				flusher.Flush()
			}
		}
	}
}

// HandleTaskDiffSSE streams git diff updates for a specific task.
func (h *Handlers) HandleTaskDiffSSE(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
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

	// Subscribe to events
	ch := h.events.Subscribe()
	defer h.events.Unsubscribe(ch)

	// Helper to send current diff state
	sendDiff := func() {
		task, err := h.tasks.Get(ctx, taskID)
		if err != nil {
			return
		}
		var html string
		if task.Status == "in_progress" {
			html = `<div class="flex flex-col items-center justify-center h-48 text-gray-500 space-y-4"><i class="fas fa-cog fa-spin text-3xl opacity-50"></i><p class="text-xs font-mono">Generating changes...</p></div>`
		} else if task.GitDiff != "" {
			html = renderDiffHTML(task.GitDiff)
		} else {
			html = `<div class="text-gray-500 italic">No changes made</div>`
		}
		fmt.Fprintf(w, "event: diff\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
		flusher.Flush()
	}

	// Send initial state
	sendDiff()

	// Periodic ticker as fallback
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Stream diff updates
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			// Only forward events for this task
			if event.TaskID != taskID {
				continue
			}
			// On any task event, refresh diff
			sendDiff()
		case <-ticker.C:
			sendDiff()
		}
	}
}

// HandleTaskAgentSSE streams agent output updates for a specific task.
func (h *Handlers) HandleTaskAgentSSE(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
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

	// Subscribe to events
	ch := h.events.Subscribe()
	defer h.events.Unsubscribe(ch)

	// Helper to send current agent state
	sendAgent := func() {
		task, err := h.tasks.Get(ctx, taskID)
		if err != nil {
			return
		}
		var html string
		if task.Status == "in_progress" {
			html = `<div class="flex flex-col items-center justify-center h-48 text-gray-500 space-y-4"><i class="fas fa-cog fa-spin text-3xl opacity-50"></i><p class="text-xs font-mono">Agent is working...</p></div>`
		} else if task.AgentOutput != "" {
			html = renderAgentHTML(task.AgentOutput)
		} else {
			html = `<div class="text-gray-500 italic">No agent output</div>`
		}
		fmt.Fprintf(w, "event: agent\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
		flusher.Flush()
	}

	// Send initial state
	sendAgent()

	// Periodic ticker as fallback
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Stream agent updates
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			// Only forward events for this task
			if event.TaskID != taskID {
				continue
			}
			// On any task event, refresh agent output
			sendAgent()
		case <-ticker.C:
			sendAgent()
		}
	}
}

// renderDiffHTML converts git diff text to styled HTML
func renderDiffHTML(diff string) string {
	if diff == "" {
		return `<div class="text-gray-500 italic">No changes made</div>`
	}

	var buf strings.Builder
	for _, line := range strings.Split(diff, "\n") {
		escapedLine := escapeHTML(line)
		if strings.HasPrefix(line, "+") {
			buf.WriteString(fmt.Sprintf(`<div class="diff-add">%s</div>`, escapedLine))
		} else if strings.HasPrefix(line, "-") {
			buf.WriteString(fmt.Sprintf(`<div class="diff-del">%s</div>`, escapedLine))
		} else {
			buf.WriteString(fmt.Sprintf(`<div>%s</div>`, escapedLine))
		}
	}
	return buf.String()
}

// renderAgentHTML converts agent output text to styled HTML
func renderAgentHTML(output string) string {
	if output == "" {
		return `<div class="text-gray-500 italic text-xs">No agent output</div>`
	}
	escaped := escapeHTML(output)
	return fmt.Sprintf(`<div class="prose prose-invert prose-xs max-w-none"><div class="text-gray-300 whitespace-pre-wrap leading-relaxed font-mono text-xs">%s</div></div>`, escaped)
}

// escapeHTML escapes HTML special characters
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
