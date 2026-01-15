package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/utils"
	"github.com/revrost/code/counterspell/internal/views"
	"github.com/revrost/code/counterspell/internal/views/components"
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
	_, _ = fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
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

			_, _ = fmt.Fprintf(w, "event: task\ndata: %s\n\n", string(data))
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
			switch t.Status {
			case "in_progress":
				active = append(active, uiTask)
			case "review", "human_review":
				uiTask.Progress = 100
				reviews = append(reviews, uiTask)
			}
		}

		// Render active rows to buffer
		var buf bytes.Buffer
		_ = views.ActiveRows(active, projects).Render(ctx, &buf)

		// Add OOB swap for reviews section
		buf.WriteString(`<div id="reviews-container" hx-swap-oob="true">`)
		_ = views.ReviewsSection(views.FeedData{Reviews: reviews, Projects: projects}).Render(ctx, &buf)
		buf.WriteString(`</div>`)

		// SSE requires single-line data, escape newlines
		html := strings.ReplaceAll(buf.String(), "\n", "")
		_, _ = fmt.Fprintf(w, "event: active-update\ndata: %s\n\n", html)
		flusher.Flush()
	}

	// Send initial state
	sendActiveUpdate()

	// Keepalive ticker - prevents proxy timeouts (ngrok, cloudflare, etc.)
	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

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
		case <-keepalive.C:
			// SSE comment to keep connection alive
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// HandleTaskSSE streams all updates (agent, diff, logs) for a specific task via a single SSE connection.
// Events: "agent", "diff", "log", "complete"
// The connection stays alive even for non-in_progress tasks to handle chat continuations.
func (h *Handlers) HandleTaskSSE(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	ctx := r.Context()

	// Check if task exists
	if _, err := h.tasks.Get(ctx, taskID); err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

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

	// Send initial state for all three types
	sendAgent := func() {
		task, err := h.tasks.Get(ctx, taskID)
		if err != nil {
			return
		}

		// For in-progress tasks, check if we have cached live history
		// This ensures SSE reconnections get the latest state even before DB persistence
		var html string
		if task.Status == models.StatusInProgress {
			if liveHistory := h.events.GetLiveHistory(taskID); liveHistory != "" {
				// Use cached live history
				html = renderAgentConversationFromJSON(ctx, liveHistory, true)
			} else {
				// Fall back to DB (might be empty for fresh tasks)
				html = renderAgentConversation(ctx, task)
			}
		} else {
			html = renderAgentConversation(ctx, task)
		}

		_, _ = fmt.Fprintf(w, "event: agent\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
		flusher.Flush()
	}

	sendDiff := func() {
		task, err := h.tasks.Get(ctx, taskID)
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
		_, _ = fmt.Fprintf(w, "event: diff\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
		flusher.Flush()
	}

	// Send initial state
	sendAgent()
	sendDiff()
	flusher.Flush()

	// Keepalive ticker - prevents proxy timeouts
	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

	// Stream all updates for this task
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

			switch event.Type {
			case "task_created":
				// Task restarted (e.g., from chat continuation) - send updated state
				
				sendAgent()
				sendDiff()

			case "agent_update":
				// Live agent update with message history
				html := renderAgentConversationFromJSON(ctx, event.HTMLPayload, true)
				_, _ = fmt.Fprintf(w, "event: agent\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
				flusher.Flush()

			case "log":
				// Log entry
				htmlData := strings.ReplaceAll(event.HTMLPayload, "\n", "")
				logHTML := fmt.Sprintf(`<div class="ml-4 relative"><div class="absolute -left-[21px] top-1 h-2.5 w-2.5 rounded-full border border-[#0D1117] bg-blue-500"></div><p class="text-xs text-gray-400">%s</p></div>`, htmlData)
				_, _ = fmt.Fprintf(w, "event: log\ndata: %s\n\n", logHTML)
				flusher.Flush()

			case "status_change":
				// Task status changed - send final state
				
				sendAgent()
				sendDiff()
				//nolint:errcheck // SSE write may fail if client disconnects
				fmt.Fprintf(w, "event: complete\ndata: {\"status\": \"%s\"}\n\n", event.HTMLPayload)
				flusher.Flush()
				// Don't close connection - keep alive for potential chat continuations

			case "todo":
				// Todo list update - forward JSON directly
				//nolint:errcheck // SSE write may fail if client disconnects
				fmt.Fprintf(w, "event: todo\ndata: %s\n\n", strings.ReplaceAll(event.HTMLPayload, "\n", ""))
				flusher.Flush()
			}

		case <-keepalive.C:
			//nolint:errcheck // SSE keepalive may fail if client disconnects
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// HandleTaskLogsSSE streams log updates for a specific task.
// Deprecated: Use HandleTaskSSE instead for unified streaming.
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
	//nolint:errcheck // SSE write may fail if client disconnects
	fmt.Fprintf(w, "event: ping\ndata: connected\n\n")
	flusher.Flush()

	// Keepalive ticker
	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

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
				_, _ = fmt.Fprintf(w, "event: log\ndata: %s\n\n", logHTML)
				flusher.Flush()
			}
		case <-keepalive.C:
			//nolint:errcheck // SSE keepalive may fail if client disconnects
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
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
		_, _ = fmt.Fprintf(w, "event: diff\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
		flusher.Flush()
	}

	// Send initial state
	sendDiff()

	// Keepalive ticker
	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

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
		case <-keepalive.C:
			//nolint:errcheck // SSE keepalive may fail if client disconnects
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
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

	// Helper to send current agent state from DB (used for initial load)
	sendAgentFromDB := func() {
		task, err := h.tasks.Get(ctx, taskID)
		if err != nil {
			return
		}
		html := renderAgentConversation(ctx, task)
		_, _ = fmt.Fprintf(w, "event: agent\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
		flusher.Flush()
	}

	// Helper to render and send agent state from JSON message history
	sendAgentFromJSON := func(messageHistoryJSON string, isInProgress bool) {
		html := renderAgentConversationFromJSON(ctx, messageHistoryJSON, isInProgress)
		_, _ = fmt.Fprintf(w, "event: agent\ndata: %s\n\n", strings.ReplaceAll(html, "\n", ""))
		flusher.Flush()
	}

	// Send initial state
	sendAgentFromDB()

	// Keepalive ticker
	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

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
			// Handle live agent updates (message history JSON in HTMLPayload)
			switch event.Type {
			case "agent_update":
				sendAgentFromJSON(event.HTMLPayload, true)
			case "status_change":
				// Task completed, reload from DB to get final state
				sendAgentFromDB()
			}
		case <-keepalive.C:
			//nolint:errcheck // SSE keepalive may fail if client disconnects
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// renderDiffHTML converts git diff text to GitHub-styled HTML
func renderDiffHTML(diff string) string {
	if diff == "" {
		return `<div class="text-gray-500 italic">No changes made</div>`
	}

	var buf strings.Builder
	var currentFile string
	var lineNum int
	inFileBlock := false

	for _, line := range strings.Split(diff, "\n") {
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
			buf.WriteString(fmt.Sprintf(`<div class="diff-file-header"><i class="fas fa-file-code mr-2 text-gray-500"></i>%s</div><div class="diff-file-body">`, escapeHTML(currentFile)))
		} else if strings.HasPrefix(line, "@@") {
			// Hunk header
			buf.WriteString(fmt.Sprintf(`<div class="diff-hunk">%s</div>`, escapedLine))
		} else if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "index ") {
			// Skip meta lines
		} else if strings.HasPrefix(line, "+") {
			lineNum++
			buf.WriteString(fmt.Sprintf(`<div class="diff-line diff-add"><span class="diff-line-num">%d</span><span class="diff-line-content">%s</span></div>`, lineNum, escapedLine))
		} else if strings.HasPrefix(line, "-") {
			buf.WriteString(fmt.Sprintf(`<div class="diff-line diff-del"><span class="diff-line-num"></span><span class="diff-line-content">%s</span></div>`, escapedLine))
		} else if line != "" {
			lineNum++
			buf.WriteString(fmt.Sprintf(`<div class="diff-line diff-context"><span class="diff-line-num">%d</span><span class="diff-line-content">%s</span></div>`, lineNum, escapedLine))
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
