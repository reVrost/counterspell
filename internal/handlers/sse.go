package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/utils"
	"github.com/revrost/code/counterspell/internal/views"
	"github.com/revrost/code/counterspell/internal/views/components"
)

// HandleSSE handles Server-Sent Events for real-time updates.
// Supports filtering by task_id via query parameter.
// If task_id is provided, only sends events for that task.
// Otherwise, sends all events (used by feed page).
func (h *Handlers) HandleSSE(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")

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

	// Get user services for rendering (only needed if task_id is specified)
	var svc *UserServices
	var err error
	ctx := r.Context()
	if taskID != "" {
		svc, err = h.getServices(ctx)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Check if task exists
		if _, err := svc.Tasks.Get(ctx, taskID); err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		// Send initial agent conversation
		lastSentID = h.sendInitialAgentState(w, flusher, ctx, svc, taskID)
		// Send initial diff
		h.sendInitialDiffState(w, flusher, ctx, svc, taskID)
	} else {
		// Feed page: send initial ping
		_, _ = fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
		flusher.Flush()
	}

	// Keepalive ticker - prevents proxy timeouts
	keepalive := time.NewTicker(10 * time.Second)
	defer keepalive.Stop()

	// Stream events
	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}

			// Filter by task_id if specified
			if taskID != "" && event.TaskID != taskID {
				continue
			}

			// Skip if we've already sent this event (deduplication)
			if event.ID <= lastSentID && event.ID != 0 {
				continue
			}
			lastSentID = event.ID

			// Send event based on type
			h.sendSSEEvent(w, flusher, ctx, svc, taskID, event)

		case <-keepalive.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (h *Handlers) sendInitialAgentState(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, svc *UserServices, taskID string) int64 {
	task, err := svc.Tasks.Get(ctx, taskID)
	if err != nil {
		return 0
	}

	var html string
	if liveHistory := h.events.GetLiveHistory(taskID); liveHistory != "" {
		html = renderAgentConversationFromJSON(ctx, liveHistory, true)
	} else if task.Status == models.StatusInProgress {
		html = renderAgentConversation(ctx, task)
	} else {
		html = renderAgentConversation(ctx, task)
	}

	eventID := h.events.GetLastEventID(taskID)
	if eventID == 0 {
		eventID = 1
	}
	_, _ = fmt.Fprintf(w, "id: %d\nevent: agent_update\ndata: %s\n\n", eventID, strings.ReplaceAll(html, "\n", ""))
	flusher.Flush()
	return eventID
}

func (h *Handlers) sendInitialDiffState(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, svc *UserServices, taskID string) {
	task, err := svc.Tasks.Get(ctx, taskID)
	if err != nil {
		return
	}

	var html string
	if task.Status == models.StatusInProgress {
		html = `<div class="flex flex-col items-center justify-center h-48 text-gray-500 space-y-4"><i class="fas fa-cog fa-spin text-3xl opacity-50"></i><p class="text-xs font-mono">Generating changes...</p></div>`
	} else if task.GitDiff != "" {
		html = renderDiffHTML(task.GitDiff)
	} else {
		html = `<div class="text-gray-500 italic">No changes made</div>`
	}

	_, _ = fmt.Fprintf(w, "event: diff_update\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
	flusher.Flush()
}

func (h *Handlers) sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, ctx context.Context, svc *UserServices, taskID string, event models.Event) {
	switch event.Type {
	case models.EventTypeAgentUpdate:
		html := renderAgentConversationFromJSON(ctx, event.HTMLPayload, true)
		_, _ = fmt.Fprintf(w, "id: %d\nevent: agent_update\ndata: %s\n\n", event.ID, strings.ReplaceAll(html, "\n", ""))
		flusher.Flush()

	case models.EventTypeLog:
		htmlData := strings.ReplaceAll(event.HTMLPayload, "\n", "")
		logHTML := fmt.Sprintf(`<div class="ml-4 relative"><div class="absolute -left-[21px] top-1 h-2.5 w-2.5 rounded-full border border-[#0D1117] bg-blue-500"></div><p class="text-xs text-gray-400">%s</p></div>`, htmlData)
		_, _ = fmt.Fprintf(w, "id: %d\nevent: log\ndata: %s\n\n", event.ID, logHTML)
		flusher.Flush()

	case models.EventTypeTodo:
		_, _ = fmt.Fprintf(w, "id: %d\nevent: todo\ndata: %s\n\n", event.ID, strings.ReplaceAll(event.HTMLPayload, "\n", ""))
		flusher.Flush()

	case models.EventTypeStatusChange:
		if taskID != "" && svc != nil {
			// Task detail page: send full state update
			h.sendInitialAgentState(w, flusher, ctx, svc, taskID)
			h.sendInitialDiffState(w, flusher, ctx, svc, taskID)

			task, _ := svc.Tasks.Get(ctx, taskID)
			statusHTML := renderStatusIndicator(task)
			_, _ = fmt.Fprintf(w, "id: %d\nevent: status\ndata: %s\n\n", event.ID, strings.ReplaceAll(statusHTML, "\n", ""))
			flusher.Flush()

			_, _ = fmt.Fprintf(w, "id: %d\nevent: complete\ndata: {\"status\": \"%s\"}\n\n", event.ID, event.HTMLPayload)
			flusher.Flush()
		} else {
			// Feed page: send JSON event for status change dont do anything la
			data, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(w, "event: task\ndata: %s\n\n", string(data))
			flusher.Flush()
		}

	case models.EventTypeTaskCreated:
		if taskID != "" && svc != nil {
			h.sendInitialAgentState(w, flusher, ctx, svc, taskID)
			h.sendInitialDiffState(w, flusher, ctx, svc, taskID)
		} else {
			data, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(w, "event: task\ndata: %s\n\n", string(data))
			flusher.Flush()
		}
	}
}

// HandleFeedActiveSSE streams active task updates via SSE for htmx.
// This replaces polling with server-push updates.
// DEPRECATED: Now using unified /events endpoint
// func (h *Handlers) _HandleFeedActiveSSE_Deprecated(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
//
// 	// Get user services upfront
// 	svc, err := h.getServices(ctx)
// 	if err != nil {
// 		http.Error(w, "Internal error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Set SSE headers
// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Connection", "keep-alive")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
//
// 	flusher, ok := w.(http.Flusher)
// 	if !ok {
// 		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Subscribe to task events
// 	ch := h.events.Subscribe()
// 	defer h.events.Unsubscribe(ch)
//
// 	// Cache projects at connection start - only refresh on project events
// 	internalProjects, _ := svc.GitHub.GetProjects(ctx)
// 	projects := make(map[string]views.UIProject)
// 	for _, p := range internalProjects {
// 		projects[p.ID] = toUIProject(p)
// 	}
//
// 	// Helper to render and send active tasks and reviews
// 	sendActiveUpdate := func() {
// 		// Get tasks from DB and categorize
// 		dbTasks, _ := svc.Tasks.List(ctx, nil, nil)
// 		var active []*views.UITask
// 		var reviews []*views.UITask
// 		for _, t := range dbTasks {
// 			uiTask := &views.UITask{
// 				ID:          t.ID,
// 				ProjectID:   t.ProjectID,
// 				Description: t.Title,
// 				AgentName:   "Agent",
// 				Status:      t.Status,
// 				Progress:    50,
// 			}
// 			switch t.Status {
// 			case models.StatusInProgress:
// 				active = append(active, uiTask)
// 			case models.StatusReview:
// 				uiTask.Progress = 100
// 				reviews = append(reviews, uiTask)
// 			}
// 		}
//
// 		// Render active rows to buffer
// 		var buf bytes.Buffer
// 		_ = views.ActiveRows(active, projects).Render(ctx, &buf)
//
// 		// Add OOB swap for reviews section
// 		buf.WriteString(`<div id="reviews-container" hx-swap-oob="true">`)
// 		_ = views.ReviewsSection(views.FeedData{Reviews: reviews, Projects: projects}).Render(ctx, &buf)
// 		buf.WriteString(`</div>`)
//
// 		// SSE requires single-line data, escape newlines
// 		html := strings.ReplaceAll(buf.String(), "\n", "")
// 		_, _ = fmt.Fprintf(w, "event: active-update\ndata: %s\n\n", html)
// 		flusher.Flush()
// 	}
//
// 	// Send initial state
// 	sendActiveUpdate()
//
// 	// Keepalive ticker - prevents proxy timeouts (ngrok, cloudflare, etc.)
// 	// 10s is safe for most proxies and ensures connection stays alive
// 	keepalive := time.NewTicker(10 * time.Second)
// 	defer keepalive.Stop()
//
// 	// Stream updates
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case event, ok := <-ch:
// 			if !ok {
// 				return
// 			}
// 			// Refresh projects cache on project-related events
// 			if event.Type == models.EventTypeProjectCreated || event.Type == models.EventTypeProjectUpdated || event.Type == models.EventTypeProjectDeleted {
// 				internalProjects, _ = svc.GitHub.GetProjects(ctx)
// 				projects = make(map[string]views.UIProject)
// 				for _, p := range internalProjects {
// 					projects[p.ID] = toUIProject(p)
// 				}
// 			}
// 			// Task event received, send update
// 			sendActiveUpdate()
// 		case <-keepalive.C:
// 			// SSE comment to keep connection alive
// 			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
// 			flusher.Flush()
// 		}
// 	}
// }

// HandleTaskSSE streams all updates (agent, diff, logs) for a specific task via a single SSE connection.
// Events: "agent", "diff", "log", "complete", "todo"
// The connection stays alive even for non-in_progress tasks to handle chat continuations.
// Supports reconnection via Last-Event-ID header for missed event replay.
// DEPRECATED: Now using unified /events endpoint
// func (h *Handlers) _HandleTaskSSE_Deprecated(w http.ResponseWriter, r *http.Request) {
// 	taskID := chi.URLParam(r, "id")
// 	ctx := r.Context()
//
// 	// Get user services upfront
// 	svc, err := h.getServices(ctx)
// 	if err != nil {
// 		http.Error(w, "Internal error", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Check if task exists
// 	if _, err := svc.Tasks.Get(ctx, taskID); err != nil {
// 		http.Error(w, "Task not found", http.StatusNotFound)
// 		return
// 	}
//
// 	// Set SSE headers
// 	w.Header().Set("Content-Type", "text/event-stream")
// 	w.Header().Set("Cache-Control", "no-cache")
// 	w.Header().Set("Connection", "keep-alive")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
//
// 	flusher, ok := w.(http.Flusher)
// 	if !ok {
// 		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
// 		return
// 	}
//
// 	// Subscribe to events
// 	ch := h.events.Subscribe()
// 	defer h.events.Unsubscribe(ch)
//
// 	// Track last sent event ID for client-side deduplication
// 	var lastSentID int64
//
// 	// Send initial state for all three types
// 	sendAgent := func() int64 {
// 		task, err := svc.Tasks.Get(ctx, taskID)
// 		if err != nil {
// 			return lastSentID
// 		}
//
// 		// Check for cached live history first (regardless of DB status)
// 		// This ensures user messages show immediately even before DB catches up
// 		var html string
// 		if liveHistory := h.events.GetLiveHistory(taskID); liveHistory != "" {
// 			// Use cached live history - always prefer this for responsiveness
// 			slog.Info("[SSE] sendAgent using live history cache", "task_id", taskID, "history_len", len(liveHistory))
// 			html = renderAgentConversationFromJSON(ctx, liveHistory, true)
// 		} else if task.Status == models.StatusInProgress {
// 			// Fall back to DB for in-progress tasks
// 			html = renderAgentConversation(ctx, task)
// 		} else {
// 			html = renderAgentConversation(ctx, task)
// 		}
//
// 		eventID := h.events.GetLastEventID(taskID)
// 		if eventID == 0 {
// 			eventID = lastSentID + 1
// 		}
// 		//nolint:errcheck // SSE write may fail if client disconnects
// 		fmt.Fprintf(w, "id: %d\nevent: agent\ndata: %s\n\n", eventID, strings.ReplaceAll(html, "\n", ""))
// 		flusher.Flush()
// 		return eventID
// 	}
//
// 	sendDiff := func() {
// 		task, err := svc.Tasks.Get(ctx, taskID)
// 		if err != nil {
// 			return
// 		}
// 		var html string
// 		if task.Status == models.StatusInProgress {
// 			html = `<div class="flex flex-col items-center justify-center h-48 text-gray-500 space-y-4"><i class="fas fa-cog fa-spin text-3xl opacity-50"></i><p class="text-xs font-mono">Generating changes...</p></div>`
// 		} else if task.GitDiff != "" {
// 			html = renderDiffHTML(task.GitDiff)
// 		} else {
// 			html = `<div class="text-gray-500 italic">No changes made</div>`
// 		}
// 		//nolint:errcheck // SSE write may fail if client disconnects
// 		fmt.Fprintf(w, "event: diff\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
// 		flusher.Flush()
// 	}
//
// 	// Send initial state
// 	lastSentID = sendAgent()
// 	sendDiff()
// 	flusher.Flush()
//
// 	// Keepalive ticker - prevents proxy timeouts
// 	keepalive := time.NewTicker(10 * time.Second)
// 	defer keepalive.Stop()
//
// 	// Stream all updates for this task
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case event, ok := <-ch:
// 			if !ok {
// 				return
// 			}
//
// 			// Only forward events for this task
// 			if event.TaskID != taskID {
// 				continue
// 			}
//
// 			// Skip if we've already sent this event (deduplication)
// 			if event.ID <= lastSentID && event.ID != 0 {
// 				continue
// 			}
// 			lastSentID = event.ID
//
// 			switch event.Type {
// 			case models.EventTypeTaskCreated:
// 				// Task restarted (e.g., from chat continuation) - send updated state
// 				slog.Info("[SSE] Received task_created, calling sendAgent", "task_id", taskID, "event_id", event.ID)
// 				lastSentID = sendAgent()
// 				sendDiff()
//
// 			case models.EventTypeAgentUpdate:
// 				// Live agent update with message history
// 				slog.Info("[SSE] Received agent_update, rendering and sending", "task_id", taskID, "event_id", event.ID, "payload_len", len(event.HTMLPayload))
// 				html := renderAgentConversationFromJSON(ctx, event.HTMLPayload, true)
// 				//nolint:errcheck // SSE write may fail if client disconnects
// 				fmt.Fprintf(w, "id: %d\nevent: agent\ndata: %s\n\n", event.ID, strings.ReplaceAll(html, "\n", ""))
// 				flusher.Flush()
// 				slog.Info("[SSE] agent_update sent to client", "task_id", taskID)
//
// 			case models.EventTypeLog:
// 				// Log entry
// 				htmlData := strings.ReplaceAll(event.HTMLPayload, "\n", "")
// 				logHTML := fmt.Sprintf(`<div class="ml-4 relative"><div class="absolute -left-[21px] top-1 h-2.5 w-2.5 rounded-full border border-[#0D1117] bg-blue-500"></div><p class="text-xs text-gray-400">%s</p></div>`, htmlData)
// 				//nolint:errcheck // SSE write may fail if client disconnects
// 				fmt.Fprintf(w, "id: %d\nevent: log\ndata: %s\n\n", event.ID, logHTML)
// 				flusher.Flush()
//
// 			case models.EventTypeStatusChange:
// 				// Task status changed - send final state
// 				lastSentID = sendAgent()
// 				sendDiff()
//
// 				// Send status indicator update (server-rendered)
// 				task, _ := svc.Tasks.Get(ctx, taskID)
// 				statusHTML := renderStatusIndicator(task)
// 				//nolint:errcheck // SSE write may fail if client disconnects
// 				fmt.Fprintf(w, "id: %d\nevent: status\ndata: %s\n\n", event.ID, strings.ReplaceAll(statusHTML, "\n", ""))
// 				flusher.Flush()
//
// 				//nolint:errcheck // SSE write may fail if client disconnects
// 				fmt.Fprintf(w, "id: %d\nevent: complete\ndata: {\"status\": \"%s\"}\n\n", event.ID, event.HTMLPayload)
// 				flusher.Flush()
// 				// Don't close connection - keep alive for potential chat continuations
//
// 			case models.EventTypeTodo:
// 				// Todo list update - forward JSON directly
// 				//nolint:errcheck // SSE write may fail if client disconnects
// 				fmt.Fprintf(w, "id: %d\nevent: todo\ndata: %s\n\n", event.ID, strings.ReplaceAll(event.HTMLPayload, "\n", ""))
// 				flusher.Flush()
// 			}
//
// 		case <-keepalive.C:
// 			//nolint:errcheck // SSE keepalive may fail if client disconnects
// 			fmt.Fprintf(w, ": keepalive\n\n")
// 			flusher.Flush()
// 		}
// 	}
// }

// renderDiffHTML converts git diff text to GitHub-styled HTML
func renderDiffHTML(diff string) string {
	if diff == "" {
		return `<div class="text-gray-500 italic">No changes made</div>`
	}

	var buf strings.Builder
	var currentFile string
	var lineNum int
	inFileBlock := false

	for line := range strings.SplitSeq(diff, "\n") {
		escapedLine := escapeHTML(line)

		if strings.HasPrefix(line, "diff --git") {
			// Close previous file block
			if inFileBlock {
				buf.WriteString(`</div>`)
			}
			inFileBlock = true
			lineNum = 0

			// Extract filename
			parts := strings.Split(line, " b/")
			if len(parts) > 1 {
				currentFile = parts[len(parts)-1]
			}
			fmt.Fprintf(&buf, `<div class="diff-file-header"><i class="fas fa-file-code mr-2 text-gray-500"></i>%s</div><div class="diff-file-body">`, escapeHTML(currentFile))
		} else if strings.HasPrefix(line, "@@") {
			// Hunk header
			fmt.Fprintf(&buf, `<div class="diff-hunk">%s</div>`, escapedLine)
		} else if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "index ") {
			// Skip meta lines
		} else if strings.HasPrefix(line, "+") {
			lineNum++
			fmt.Fprintf(&buf, `<div class="diff-line diff-add"><span class="diff-line-num">%d</span><span class="diff-line-content">%s</span></div>`, lineNum, escapedLine)
		} else if strings.HasPrefix(line, "-") {
			fmt.Fprintf(&buf, `<div class="diff-line diff-del"><span class="diff-line-num"></span><span class="diff-line-content">%s</span></div>`, escapedLine)
		} else if line != "" {
			lineNum++
			fmt.Fprintf(&buf, `<div class="diff-line diff-context"><span class="diff-line-num">%d</span><span class="diff-line-content">%s</span></div>`, lineNum, escapedLine)
		}
	}

	if inFileBlock {
		buf.WriteString(`</div>`)
	}

	return buf.String()
}

// renderAgentHTML converts agent output text to styled HTML with markdown support
func renderAgentHTML(output string) string {
	if output == "" {
		return `<div class="text-gray-500 italic text-xs">No agent output</div>`
	}
	html := utils.RenderMarkdownHTML(output)
	return fmt.Sprintf(`<div class="text-sm text-gray-300 leading-normal prose prose-invert prose-sm prose-p:my-3 prose-headings:font-bold prose-headings:text-sm prose-headings:mt-4 prose-headings:mb-2 prose-code:text-xs prose-pre:text-xs prose-pre:my-3 prose-ul:my-2 prose-ol:my-2 prose-li:my-1 max-w-none">%s</div>`, html)
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

// renderAgentConversation renders the full conversation history using templ components
func renderAgentConversation(ctx context.Context, task *models.Task) string {
	// Parse message history
	var uiMessages []views.UIMessage
	if task.MessageHistory != "" {
		var rawMessages []struct {
			Role    string `json:"role"`
			Content []struct {
				Type      string         `json:"type"`
				Text      string         `json:"text,omitempty"`
				Name      string         `json:"name,omitempty"`
				Input     map[string]any `json:"input,omitempty"`
				ID        string         `json:"id,omitempty"`
				ToolUseID string         `json:"tool_use_id,omitempty"`
				Content   string         `json:"content,omitempty"`
			} `json:"content"`
		}
		if err := json.Unmarshal([]byte(task.MessageHistory), &rawMessages); err == nil {
			for _, msg := range rawMessages {
				uiMsg := views.UIMessage{Role: msg.Role}
				for _, block := range msg.Content {
					uiContent := views.UIContent{Type: block.Type}
					switch block.Type {
					case "text":
						uiContent.Text = block.Text
					case "tool_use":
						uiContent.ToolName = block.Name
						uiContent.ToolID = block.ID
						if inputBytes, err := json.Marshal(block.Input); err == nil {
							uiContent.ToolInput = string(inputBytes)
						}
					case "tool_result":
						uiContent.ToolID = block.ToolUseID
						uiContent.Text = block.Content
					}
					uiMsg.Content = append(uiMsg.Content, uiContent)
				}
				uiMessages = append(uiMessages, uiMsg)
			}
		}
	}

	var buf bytes.Buffer

	// Render messages if we have them
	if len(uiMessages) > 0 {
		buf.WriteString(`<div class="space-y-0">`)
		for _, msg := range uiMessages {
			_ = components.MessageBubble(msg).Render(ctx, &buf)
		}
		buf.WriteString(`</div>`)
	}

	// If in progress, show compact loading
	if task.Status == models.StatusInProgress {
		buf.WriteString(`<div class="flex items-center gap-3 px-4 py-3">`)
		buf.WriteString(`<div class="relative">`)
		buf.WriteString(`<div class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center">`)
		buf.WriteString(`<i class="fas fa-robot text-sm text-violet-400 pulse-glow"></i>`)
		buf.WriteString(`</div>`)
		buf.WriteString(`<div class="absolute inset-0 animate-spin" style="animation-duration: 3s;">`)
		buf.WriteString(`<div class="absolute -top-0.5 left-1/2 -translate-x-1/2 w-1 h-1 bg-violet-400 rounded-full"></div>`)
		buf.WriteString(`</div>`)
		buf.WriteString(`</div>`)
		buf.WriteString(`<div>`)
		buf.WriteString(`<p class="text-xs font-medium shimmer">Agent is thinking...</p>`)
		buf.WriteString(`<p class="text-[10px] text-gray-600">Analyzing code</p>`)
		buf.WriteString(`</div>`)
		buf.WriteString(`</div>`)
	} else if len(uiMessages) == 0 {
		// No messages and not in progress - fallback to raw output or empty
		if task.AgentOutput != "" {
			return renderAgentHTML(task.AgentOutput)
		}
		return `<div class="p-5 text-gray-500 italic text-xs">No agent output</div>`
	}

	return buf.String()
}

// renderStatusIndicator renders the task status indicator dot
func renderStatusIndicator(task *models.Task) string {
	if task == nil {
		return ""
	}
	switch task.Status {
	case models.StatusInProgress:
		return `<div class="w-1.5 h-1.5 rounded-full bg-orange-400" title="In Progress"></div>`
	case models.StatusReview:
		return `<div class="w-1.5 h-1.5 rounded-full bg-blue-400" title="Needs Review"></div>`
	case models.StatusDone:
		return `<div class="w-1.5 h-1.5 rounded-full bg-green-400" title="Done"></div>`
	default:
		return `<div class="w-1.5 h-1.5 rounded-full bg-gray-400" title="Unknown"></div>`
	}
}

// renderAgentConversationFromJSON renders conversation from JSON message history (for live streaming)
func renderAgentConversationFromJSON(ctx context.Context, messageHistoryJSON string, isInProgress bool) string {
	// Parse message history
	var uiMessages []views.UIMessage
	if messageHistoryJSON != "" {
		var rawMessages []struct {
			Role    string `json:"role"`
			Content []struct {
				Type      string         `json:"type"`
				Text      string         `json:"text,omitempty"`
				Name      string         `json:"name,omitempty"`
				Input     map[string]any `json:"input,omitempty"`
				ID        string         `json:"id,omitempty"`
				ToolUseID string         `json:"tool_use_id,omitempty"`
				Content   string         `json:"content,omitempty"`
			} `json:"content"`
		}
		if err := json.Unmarshal([]byte(messageHistoryJSON), &rawMessages); err == nil {
			for _, msg := range rawMessages {
				uiMsg := views.UIMessage{Role: msg.Role}
				for _, block := range msg.Content {
					uiContent := views.UIContent{Type: block.Type}
					switch block.Type {
					case "text":
						uiContent.Text = block.Text
					case "tool_use":
						uiContent.ToolName = block.Name
						uiContent.ToolID = block.ID
						if inputBytes, err := json.Marshal(block.Input); err == nil {
							uiContent.ToolInput = string(inputBytes)
						}
					case "tool_result":
						uiContent.ToolID = block.ToolUseID
						uiContent.Text = block.Content
					}
					uiMsg.Content = append(uiMsg.Content, uiContent)
				}
				uiMessages = append(uiMessages, uiMsg)
			}
		}
	}

	var buf bytes.Buffer

	// Render messages if we have them
	if len(uiMessages) > 0 {
		buf.WriteString(`<div class="space-y-0">`)
		for _, msg := range uiMessages {
			_ = components.MessageBubble(msg).Render(ctx, &buf)
		}
		buf.WriteString(`</div>`)
	}

	// If in progress, show compact loading
	if isInProgress {
		buf.WriteString(`<div class="flex items-center gap-3 px-4 py-3">`)
		buf.WriteString(`<div class="relative">`)
		buf.WriteString(`<div class="w-8 h-8 rounded-lg bg-violet-500/10 border border-violet-500/20 flex items-center justify-center">`)
		buf.WriteString(`<i class="fas fa-robot text-sm text-violet-400 pulse-glow"></i>`)
		buf.WriteString(`</div>`)
		buf.WriteString(`<div class="absolute inset-0 animate-spin" style="animation-duration: 3s;">`)
		buf.WriteString(`<div class="absolute -top-0.5 left-1/2 -translate-x-1/2 w-1 h-1 bg-violet-400 rounded-full"></div>`)
		buf.WriteString(`</div>`)
		buf.WriteString(`</div>`)
		buf.WriteString(`<div>`)
		buf.WriteString(`<p class="text-xs font-medium shimmer">Agent is thinking...</p>`)
		buf.WriteString(`<p class="text-[10px] text-gray-600">Analyzing code</p>`)
		buf.WriteString(`</div>`)
		buf.WriteString(`</div>`)
	} else if len(uiMessages) == 0 {
		return `<div class="p-5 text-gray-500 italic text-xs">No agent output</div>`
	}

	return buf.String()
}
