// Package handlers contains the handlers for the API.
package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/views"
	"github.com/revrost/code/counterspell/internal/views/components"
	"github.com/revrost/code/counterspell/internal/views/layout"
)

// HandleHome renders the public landing page for unauthenticated users
func (h *Handlers) HandleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := layout.Home().Render(ctx, w); err != nil {
		slog.Error("[HOME] Failed to render landing page", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Helper to convert internal project to UI project
func toUIProject(p models.Project) views.UIProject {
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

// RenderApp renders the base templ page
func (h *Handlers) RenderApp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.Info("[FEED] HandleFeed called", "method", r.Method, "url", r.URL.String(), "hx_request", r.Header.Get("HX-Request"))

	// Get user services
	svc, err := h.getServices(ctx)
	if err != nil {
		slog.Error("[FEED] Failed to get services", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Get Settings
	settings, err := svc.Settings.GetSettings(ctx)
	if err != nil {
		slog.Error("[FEED] Failed to get settings", "error", err)
		settings = &models.UserSettings{} // Default empty settings
	}

	// Get projects
	internalProjects, err := svc.GitHub.GetProjects(ctx)
	if err != nil {
		slog.Error("[FEED] Failed to get projects", "error", err)
	}
	slog.Info("[FEED] Loaded projects", "count", len(internalProjects))

	projects := make(map[string]views.UIProject)
	for _, p := range internalProjects {
		projects[p.ID] = toUIProject(p)
	}

	// Load real tasks from DB
	dbTasks, _ := svc.Tasks.List(ctx, nil, nil)

	data := views.FeedData{
		Projects: projects,
	}

	for _, t := range dbTasks {
		uiTask := &views.UITask{
			ID:          t.ID,
			ProjectID:   t.ProjectID,
			Description: t.Title,
			AgentName:   "Agent",
			Status:      t.Status,
			Progress:    50,
		}

		switch t.Status {
		case models.StatusTodo:
			uiTask.Progress = 0
			data.Todo = append(data.Todo, uiTask)
		case models.StatusReview:
			uiTask.Progress = 100
			data.Reviews = append(data.Reviews, uiTask)
		case models.StatusInProgress:
			data.Active = append(data.Active, uiTask)
		case models.StatusDone:
			uiTask.Progress = 100
			data.Done = append(data.Done, uiTask)
		}
	}

	// If this is an HTMX request, render only the feed component (partial)
	if r.Header.Get("HX-Request") == "true" {
		if err := views.Feed(data).Render(ctx, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Check authentication via GitHub connection
	isAuthenticated := false
	conn, connErr := svc.GitHub.GetActiveConnection(ctx)
	if connErr == nil && conn != nil {
		isAuthenticated = true
		slog.Info("[FEED] User is authenticated", "login", conn.Login, "type", conn.Type)
	} else {
		slog.Info("[FEED] User is NOT authenticated", "error", connErr)
	}

	// Get user email from JWT claims
	userEmail := ""
	if claims := auth.ClaimsFromContext(ctx); claims != nil {
		userEmail = claims.Email
	}

	slog.Info("[FEED] Rendering page", "isAuthenticated", isAuthenticated, "projectCount", len(projects))
	component := layout.Base("Counterspell", projects, *settings, isAuthenticated, userEmail, data)
	if err := component.Render(ctx, w); err != nil {
		slog.Error("[FEED] Failed to render", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleFeed returns the active rows partial
func (h *Handlers) HandleFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.Info("[FEED_ACTIVE] Rendering active rows")

	svc, err := h.getServices(ctx)
	if err != nil {
		slog.Error("[FEED_ACTIVE] Failed to get services", "error", err)
		return
	}

	// Get real projects
	internalProjects, _ := svc.GitHub.GetProjects(ctx)
	projects := make(map[string]views.UIProject)
	for _, p := range internalProjects {
		projects[p.ID] = toUIProject(p)
	}

	// Get active tasks from DB
	dbTasks, _ := svc.Tasks.List(ctx, nil, nil)
	var active []*views.UITask
	var reviews []*views.UITask
	for _, t := range dbTasks {
		uiTask := &views.UITask{
			ID:          t.ID,
			ProjectID:   t.ProjectID,
			Description: t.Title,
			AgentName:   "Agent",
			Status:      t.Status,
			Progress:    50,
		}
		switch t.Status {
		case models.StatusInProgress:
			active = append(active, uiTask)
		case models.StatusReview:
			uiTask.Progress = 100
			reviews = append(reviews, uiTask)
		}
	}

	// Render Active Rows
	// if err := views.ActiveRows(active, projects).Render(ctx, w); err != nil {
	// 	slog.Error("render error", "error", err)
	// }

	// Render Reviews OOB
	//nolint:errcheck
	w.Write([]byte(`<div id="feed-content" hx-swap-oob="true">`))
	if err := views.Feed(views.FeedData{
		Active:   active,
		Reviews:  reviews,
		Projects: projects,
	}).Render(ctx, w); err != nil {
		slog.Error("render error", "error", err)
	}
	//nolint:errcheck
	w.Write([]byte(`</div>`))
}

// HandleTaskDetailUI renders the task detail modal content
func (h *Handlers) HandleTaskDetailUI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	svc, err := h.getServices(ctx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Get real task from database
	dbTask, err := svc.Tasks.Get(ctx, id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Get the project
	projects, err := svc.GitHub.GetProjects(ctx)
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

	if project.ID == "" {
		project = views.UIProject{ID: dbTask.ProjectID, Name: dbTask.ProjectID, Icon: "fa-folder", Color: "text-gray-400"}
	}

	// Load logs from DB
	dbLogs, err := svc.Tasks.GetLogs(ctx, id)
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

	task := &views.UITask{
		ID:          dbTask.ID,
		ProjectID:   dbTask.ProjectID,
		Description: dbTask.Title,
		AgentName:   "Agent",
		Status:      dbTask.Status,
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
		// h.HandleFeed(w, r)
		return
	}

	if projectID == "" {
		w.Header().Set("HX-Trigger", `{"toast": "Select a project first", "taskCreated": "false"}`)
		h.RenderApp(w, r)
		return
	}

	orchestrator, err := h.getOrchestrator(ctx)
	if err != nil {
		slog.Error("Failed to get orchestrator", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Internal error", "taskCreated": "false"}`)
		h.RenderApp(w, r)
		return
	}

	_, err = orchestrator.StartTask(ctx, projectID, intent, modelID)
	if err != nil {
		slog.Error("Failed to start task", "error", err)
		w.Header().Set("HX-Trigger", `{"toast": "Failed to start task", "taskCreated": "false"}`)
		h.RenderApp(w, r)
		return
	}

	w.Header().Set("HX-Trigger", `{"toast": "Task started", "taskCreated": "true"}`)
	h.RenderApp(w, r)
}

// HandleSaveSettings saves user settings
func (h *Handlers) HandleSaveSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	svc, err := h.getServices(ctx)
	if err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Internal error"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := auth.UserIDFromContext(ctx)
	if userID == "" {
		userID = "default"
	}

	settings := &models.UserSettings{
		UserID:        userID,
		OpenRouterKey: r.FormValue("openrouter_key"),
		ZaiKey:        r.FormValue("zai_key"),
		AnthropicKey:  r.FormValue("anthropic_key"),
		OpenAIKey:     r.FormValue("openai_key"),
		AgentBackend:  r.FormValue("agent_backend"),
	}

	if err := svc.Settings.UpdateSettings(ctx, settings); err != nil {
		w.Header().Set("HX-Trigger", `{"toast": "Failed to save settings: `+err.Error()+`"}`)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Settings saved successfully"}`)
	w.WriteHeader(http.StatusOK)
}

// HandleFileSearch searches for files in a project using fuzzy matching.
func (h *Handlers) HandleFileSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := r.URL.Query().Get("project_id")
	query := r.URL.Query().Get("q")

	if projectID == "" {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}

	orchestrator, err := h.getOrchestrator(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
		return
	}

	files, err := orchestrator.SearchProjectFiles(ctx, projectID, query, 20)
	if err != nil {
		slog.Error("File search failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
		return
	}

	if files == nil {
		files = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		slog.Error("Failed to encode file search results", "error", err)
	}
}

// HandleTranscribe transcribes uploaded audio to text.
func (h *Handlers) HandleTranscribe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "No audio file provided", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")

	text, err := h.transcription.TranscribeAudio(ctx, file, contentType)
	if err != nil {
		slog.Error("Transcription failed", "error", err)
		http.Error(w, "Transcription failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//nolint:errcheck
	w.Write([]byte(text))
}
