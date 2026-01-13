package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
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

	// Get Settings
	settings, _ := h.settings.GetSettings(ctx)

	// Get projects
	internalProjects, err := h.github.GetProjects(ctx)
	if err != nil {
		slog.Error("Failed to get projects", "error", err)
	}
	slog.Info("Loaded projects", "count", len(internalProjects))

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
	if conn, err := h.github.GetActiveConnection(ctx); err == nil && conn != nil {
		isAuthenticated = true
	}

	component := layout.Base("Counterspell", projects, *settings, isAuthenticated, views.Feed(data))
	if err := component.Render(ctx, w); err != nil {
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
		if t.Status == "in_progress" {
			active = append(active, uiTask)
		} else if t.Status == "review" || t.Status == "human_review" {
			uiTask.Progress = 100
			reviews = append(reviews, uiTask)
		}
	}

	// Render Active Rows
	views.ActiveRows(active, projects).Render(ctx, w)

	// Render Reviews OOB
	w.Write([]byte(`<div id="reviews-container" hx-swap-oob="true">`))
	views.ReviewsSection(views.FeedData{Reviews: reviews, Projects: projects}).Render(ctx, w)
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
	}

	components.TaskDetail(task, project).Render(ctx, w)
}

// HandleActionRetry mocks retry action
func (h *Handlers) HandleActionRetry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Task restarting..."}`)
	h.HandleFeed(w, r)
}

// HandleActionMerge mocks merge action
func (h *Handlers) HandleActionMerge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Merged successfully"}`)
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

	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Continuing task..."}`)
	h.HandleFeed(w, r)
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
