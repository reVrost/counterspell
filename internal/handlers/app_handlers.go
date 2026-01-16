package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/git"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/views"
	"github.com/revrost/code/counterspell/internal/views/components"
	"github.com/revrost/code/counterspell/internal/views/layout"
)

// Helper to convert internal project to UI project
func toUIProject(p models.Project) views.UIProject {
	// Generate a deterministic color/icon based on ID or Name
	colors := []string{"text-blue-400", "text-purple-400", "text-green-400", "text-yellow-400", "text-pink-400"}
	icons := []string{"fa-server", "fa-columns", "fa-mobile-alt", "fa-database", "fa-globe"}

	idx := 0
	for i, c := range p.ID {
		idx += int(c) * (i + 1)
	}

	return views.UIProject{
		ID:    p.ID,
		Name:  fmt.Sprintf("%s/%s", p.GitHubOwner, p.GitHubRepo),
		Icon:  icons[idx%len(icons)],
		Color: colors[idx%len(colors)],
	}
}

// HandleFeed renders the main feed page
func (h *Handlers) HandleFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.Info("[FEED] HandleFeed called", "method", r.Method, "url", r.URL.String(), "hx_request", r.Header.Get("HX-Request"))

	// Get Settings
	settings, err := h.settings.GetSettings(ctx)
	if err != nil {
		slog.Error("[FEED] Failed to get settings", "error", err)
	}

	// Get projects
	internalProjects, err := h.github.GetProjects(ctx)
	if err != nil {
		slog.Error("[FEED] Failed to get projects", "error", err)
	}
	slog.Info("[FEED] Loaded projects", "count", len(internalProjects))

	projects := make(map[string]views.UIProject)
	for _, p := range internalProjects {
		projects[p.ID] = toUIProject(p)
	}

	// Load real tasks from DB
	dbTasks, _ := h.tasks.List(ctx, nil, nil)

	data := views.FeedData{
		Projects: projects,
	}

	for _, t := range dbTasks {
		uiTask := &views.UITask{
			ID:          t.ID,
			ProjectID:   t.ProjectID,
			Description: t.Title,
			AgentName:   "Agent",
			Status:      string(t.Status),
			Progress:    50, // TODO: track real progress
		}

		switch t.Status {
		case "todo":
			uiTask.Progress = 0
			data.Todo = append(data.Todo, uiTask)
		case "review", "human_review":
			uiTask.Progress = 100
			data.Reviews = append(data.Reviews, uiTask)
		case "in_progress":
			data.Active = append(data.Active, uiTask)
		case "done":
			uiTask.Progress = 100
			data.Done = append(data.Done, uiTask)
		}
	}



	// If this is an HTMX request, render only the feed component (partial)
	// Otherwise, render the full page layout
	if r.Header.Get("HX-Request") == "true" {
		if err := views.Feed(data).Render(ctx, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Check authentication via GitHub connection
	isAuthenticated := false
	conn, connErr := h.github.GetActiveConnection(ctx)
	if connErr == nil && conn != nil {
		isAuthenticated = true
		slog.Info("[FEED] User is authenticated", "login", conn.Login, "type", conn.Type)
	} else {
		slog.Info("[FEED] User is NOT authenticated", "error", connErr)
	}

	slog.Info("[FEED] Rendering page", "isAuthenticated", isAuthenticated, "projectCount", len(projects))
	component := layout.Base("Counterspell", projects, *settings, isAuthenticated, views.Feed(data))
	if err := component.Render(ctx, w); err != nil {
		slog.Error("[FEED] Failed to render", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleFeedActive returns the active rows partial
func (h *Handlers) HandleFeedActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get real projects
	internalProjects, _ := h.github.GetProjects(ctx)
	projects := make(map[string]views.UIProject)
	for _, p := range internalProjects {
		projects[p.ID] = toUIProject(p)
	}

	// Get active tasks from DB
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

	// Render Active Rows
	if err := views.ActiveRows(active, projects).Render(ctx, w); err != nil {
		slog.Error("render error", "error", err)
	}

	// Render Reviews OOB (writes may fail if client disconnects - expected)
	//nolint:errcheck
	w.Write([]byte(`<div id="reviews-container" hx-swap-oob="true">`))
	if err := views.ReviewsSection(views.FeedData{Reviews: reviews, Projects: projects}).Render(ctx, w); err != nil {
		slog.Error("render error", "error", err)
	}
	//nolint:errcheck
	w.Write([]byte(`</div>`))
}

// HandleTaskDetail renders the task detail modal content
func (h *Handlers) HandleTaskDetailUI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	// Get real task from database
	dbTask, err := h.tasks.Get(ctx, id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Get the project from the projects map
	projects, err := h.github.GetProjects(ctx)
	if err != nil {
		slog.Error("Failed to get projects", "error", err)
		http.Error(w, "Failed to load project", http.StatusInternalServerError)
		return
	}

	var project views.UIProject
	for _, p := range projects {
		if p.ID == dbTask.ProjectID {
			project = toUIProject(p)
			break
		}
	}

	// If project not found, use default values
	if project.ID == "" {
		project = views.UIProject{ID: dbTask.ProjectID, Name: dbTask.ProjectID, Icon: "fa-folder", Color: "text-gray-400"}
	}

	// Load logs from DB
	dbLogs, err := h.tasks.GetLogs(ctx, id)
	if err != nil {
		slog.Error("Failed to get logs", "error", err)
	}

	// Convert to UI logs
	var uiLogs []views.UILogEntry
	for _, log := range dbLogs {
		uiLogs = append(uiLogs, views.UILogEntry{
			Timestamp: log.CreatedAt,
			Message:   log.Message,
			Type:      string(log.Level),
		})
	}

	// Parse message history
	var uiMessages []views.UIMessage
	if dbTask.MessageHistory != "" {
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
		if err := json.Unmarshal([]byte(dbTask.MessageHistory), &rawMessages); err == nil {
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

	// Create UI task from DB task
	task := &views.UITask{
		ID:          dbTask.ID,
		ProjectID:   dbTask.ProjectID,
		Description: dbTask.Title,
		AgentName:   "Agent",
		Status:      string(dbTask.Status),
		Progress:    100,
		AgentOutput: dbTask.AgentOutput,
		GitDiff:     dbTask.GitDiff,
		PreviewURL:  "",
		Logs:        uiLogs,
		Messages:    uiMessages,
	}

	if err := components.TaskDetail(task, project).Render(ctx, w); err != nil {
		slog.Error("render error", "error", err)
	}
}

// HandleActionRetry mocks retry action
func (h *Handlers) HandleActionRetry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Task restarting..."}`)
	h.HandleFeed(w, r)
}

// HandleActionClear clears a task's chat history and context
func (h *Handlers) HandleActionClear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	if err := h.tasks.ClearHistory(ctx, taskID); err != nil {
		slog.Error("Failed to clear task history", "error", err, "task_id", taskID)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to clear history"}`)
		h.HandleFeed(w, r)
		return
	}

	w.Header().Set("HX-Trigger", `{"toast": "Chat history cleared"}`)
	h.HandleFeed(w, r)
}

// HandleActionMerge merges a task's branch to main and pushes
func (h *Handlers) HandleActionMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	err := h.orchestrator.MergeTask(ctx, taskID)
	if err != nil {
		// Check if it's a merge conflict
		if conflictErr, ok := err.(*git.ErrMergeConflict); ok {
			slog.Info("Merge conflict detected, showing conflict UI", "task_id", taskID, "files", conflictErr.ConflictedFiles)

			// Get conflict details
			conflicts, err := h.orchestrator.GetConflictDetails(ctx, taskID, conflictErr.ConflictedFiles)
			if err != nil {
				slog.Error("Failed to get conflict details", "error", err)
				w.Header().Set("HX-Trigger", `{"toast": "Failed to load conflict details"}`)
				h.HandleFeed(w, r)
				return
			}

			// Convert to component type
			uiConflicts := make([]components.ConflictFile, len(conflicts))
			for i, c := range conflicts {
				uiConflicts[i] = components.ConflictFile{
					Path:   c.Path,
					Ours:   c.Ours,
					Theirs: c.Theirs,
				}
			}

			// Render conflict view
			if err := components.ConflictView(taskID, uiConflicts).Render(ctx, w); err != nil {
				slog.Error("render error", "error", err)
			}
			return
		}

		slog.Error("Failed to merge task", "task_id", taskID, "error", err)
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"toast": "Merge failed: %s"}`, err.Error()))
		h.HandleFeed(w, r)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Merged to main and pushed!"}`)
	h.HandleFeed(w, r)
}

// HandleResolveConflict resolves a single file conflict
func (h *Handlers) HandleResolveConflict(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Invalid request"}`)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filePath := r.FormValue("file")
	choice := r.FormValue("choice")

	// Get conflict details to get the content
	conflicts, err := h.orchestrator.GetConflictDetails(ctx, taskID, []string{filePath})
	if err != nil || len(conflicts) == 0 {
		w.Header().Set("HX-Trigger", `{"toast": "Failed to get conflict"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var resolution string
	switch choice {
	case "ours":
		resolution = conflicts[0].Ours
	case "theirs":
		resolution = conflicts[0].Theirs
	default:
		w.Header().Set("HX-Trigger", `{"toast": "Invalid choice"}`)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.orchestrator.ResolveConflict(ctx, taskID, filePath, resolution); err != nil {
		slog.Error("Failed to resolve conflict", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to resolve conflict"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"toast": "Resolved %s"}`, filePath))
	w.WriteHeader(http.StatusOK)
}

// HandleAbortMerge aborts the current merge
func (h *Handlers) HandleAbortMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	if err := h.orchestrator.AbortMerge(ctx, taskID); err != nil {
		slog.Error("Failed to abort merge", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to abort merge"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Merge aborted"}`)
	h.HandleFeed(w, r)
}

// HandleCompleteMerge completes the merge after conflicts are resolved
func (h *Handlers) HandleCompleteMerge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	if err := h.orchestrator.CompleteMergeResolution(ctx, taskID); err != nil {
		slog.Error("Failed to complete merge", "error", err)
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"toast": "Merge failed: %s"}`, err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Merged to main!"}`)
	h.HandleFeed(w, r)
}

// HandleActionPR creates a GitHub Pull Request for the task
func (h *Handlers) HandleActionPR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	prURL, err := h.orchestrator.CreatePR(ctx, taskID)
	if err != nil {
		slog.Error("Failed to create PR", "task_id", taskID, "error", err)
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"toast": "Failed to create PR: %s"}`, err.Error()))
		h.HandleFeed(w, r)
		return
	}

	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"close-modal": true, "toast": "PR created!", "open-url": "%s"}`, prURL))
	h.HandleFeed(w, r)
}

// HandleActionDiscard deletes a task and cleans up its worktree
func (h *Handlers) HandleActionDiscard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	// Clean up worktree first (before DB delete)
	if err := h.orchestrator.CleanupTask(id); err != nil {
		slog.Error("Failed to cleanup worktree", "task_id", id, "error", err)
		// Continue with delete even if cleanup fails
	}

	if err := h.tasks.Delete(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Task discarded"}`)
	h.HandleFeed(w, r)
}

// HandleActionChat continues a task with a follow-up message.
func (h *Handlers) HandleActionChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := chi.URLParam(r, "id")

	if err := r.ParseForm(); err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Invalid request"}`)
		h.HandleFeed(w, r)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		w.Header().Set("HX-Trigger", `{"toast": "Message required"}`)
		h.HandleFeed(w, r)
		return
	}

	// Get default model from settings or use a default
	modelID := "o#anthropic/claude-sonnet-4" // Default model

	// Continue the task with the follow-up message
	if err := h.orchestrator.ContinueTask(ctx, taskID, message, modelID); err != nil {
		slog.Error("Failed to continue task", "task_id", taskID, "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to continue task"}`)
		h.HandleFeed(w, r)
		return
	}

	// SSE will stream updates to agent tab
	w.WriteHeader(http.StatusOK)
}

// HandleAddTask creates a new task and starts execution.
func (h *Handlers) HandleAddTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	intent := r.FormValue("voice_input")
	projectID := r.FormValue("project_id")
	modelID := r.FormValue("model_id")

	if intent == "" {
		h.HandleFeed(w, r)
		return
	}

	if projectID == "" {
		w.Header().Set("HX-Trigger", `{"toast": "Select a project first", "taskCreated": "false"}`)
		h.HandleFeed(w, r)
		return
	}

	// Start task execution
	_, err := h.orchestrator.StartTask(ctx, projectID, intent, modelID)
	if err != nil {
		slog.Error("Failed to start task", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to start task", "taskCreated": "false"}`)
		h.HandleFeed(w, r)
		return
	}

	w.Header().Set("HX-Trigger", `{"toast": "Task started", "taskCreated": "true"}`)
	h.HandleFeed(w, r)
}

// HandleSaveSettings saves user settings
func (h *Handlers) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	settings := &models.UserSettings{
		UserID:        "default",
		OpenRouterKey: r.FormValue("openrouter_key"),
		ZaiKey:        r.FormValue("zai_key"),
		AnthropicKey:  r.FormValue("anthropic_key"),
		OpenAIKey:     r.FormValue("openai_key"),
		AgentBackend:  r.FormValue("agent_backend"),
	}

	if err := h.settings.UpdateSettings(ctx, settings); err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Failed to save settings: `+err.Error()+`"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Settings saved successfully"}`)
	// We don't need to re-render settings as they are saved.
	w.WriteHeader(http.StatusOK)
}

// HandleTranscribe transcribes uploaded audio to text.
func (h *Handlers) HandleTranscribe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the audio file
	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "No audio file provided", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// Get content type for format detection
	contentType := header.Header.Get("Content-Type")

	// Transcribe
	text, err := h.transcription.TranscribeAudio(ctx, file, contentType)
	if err != nil {
		slog.Error("Transcription failed", "error", err)
		http.Error(w, "Transcription failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return plain text (write may fail if client disconnects)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//nolint:errcheck
	w.Write([]byte(text))
}
