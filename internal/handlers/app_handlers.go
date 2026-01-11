package handlers

import (
	"fmt"
	"net/http"
	"time"

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
		Icon:  icons[idx % len(icons)],
		Color: colors[idx % len(colors)],
	}
}

// HandleFeed renders the main feed page
func (h *Handlers) HandleFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get Settings
	settings, _ := h.settings.GetSettings(ctx)

	// Get projects
	internalProjects, _ := h.github.GetProjects(ctx)
	projects := make(map[string]views.UIProject)
	for _, p := range internalProjects {
		projects[p.ID] = toUIProject(p)
	}

	// Add mock projects if none
	if len(projects) == 0 {
		projects["core"] = views.UIProject{ID: "core", Name: "acme/core-platform", Icon: "fa-server", Color: "text-blue-400"}
		projects["web"] = views.UIProject{ID: "web", Name: "acme/web-dashboard", Icon: "fa-columns", Color: "text-purple-400"}
		projects["ios"] = views.UIProject{ID: "ios", Name: "acme/ios-app", Icon: "fa-mobile-alt", Color: "text-green-400"}
	}

	// Mock Tasks for UI
	tasks := []*views.UITask{
		{
			ID:          "1",
			ProjectID:   "core",
			Description: "Refactor auth middleware to support OIDC",
			AgentName:   "Agent-007",
			Status:      "done",
			Progress:    100,
		},
		{
			ID:          "2",
			ProjectID:   "web",
			Description: "Add Dark Mode to Dashboard",
			AgentName:   "Agent-042",
			Status:      "review",
			Progress:    100,
		},
		{
			ID:          "3",
			ProjectID:   "ios",
			Description: "Fix crash on startup",
			AgentName:   "Agent-101",
			Status:      "in_progress",
			Progress:    45,
		},
	}

	data := views.FeedData{
		Projects: projects,
	}

	for _, t := range tasks {
		if t.Status == "review" {
			data.Reviews = append(data.Reviews, t)
		} else if t.Status == "in_progress" || t.Status == "todo" {
			data.Active = append(data.Active, t)
		} else if t.Status == "done" {
			data.Done = append(data.Done, t)
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

	// Check authentication
	isAuthenticated := false
	if h.auth != nil {
		if _, err := h.auth.GetSession(r); err == nil {
			isAuthenticated = true
		}
	}

	component := layout.Base("Counterspell", projects, *settings, isAuthenticated, views.Feed(data))
	if err := component.Render(ctx, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleFeedActive returns the active rows partial
func (h *Handlers) HandleFeedActive(w http.ResponseWriter, r *http.Request) {
	// Mock logic: randomly increment progress of active tasks
	// In real app, we fetch from DB
	mockActive := []*views.UITask{
		{
			ID:          "3",
			ProjectID:   "ios",
			Description: "Fix crash on startup",
			AgentName:   "Agent-101",
			Status:      "in_progress",
			Progress:    int(time.Now().Unix() % 100), // Random progress
		},
	}
	
	// Mock Projects map
	projects := map[string]views.UIProject{
		"ios": {ID: "ios", Name: "acme/ios-app", Icon: "fa-mobile-alt", Color: "text-green-400"},
	}

	// Also Update Reviews via OOB
	mockReviews := []*views.UITask{
		{
			ID:          "2",
			ProjectID:   "web",
			Description: "Add Dark Mode to Dashboard",
			AgentName:   "Agent-042",
			Status:      "review",
			Progress:    100,
		},
	}
	reviewsProjects := map[string]views.UIProject{
		"web": {ID: "web", Name: "acme/web-dashboard", Icon: "fa-columns", Color: "text-purple-400"},
	}
	// Merge maps
	for k, v := range reviewsProjects { projects[k] = v }

	// Render Active Rows
	views.ActiveRows(mockActive, projects).Render(r.Context(), w)
	
	// Render Reviews OOB
	w.Write([]byte(`<div id="reviews-container">`))
	views.ReviewsSection(views.FeedData{Reviews: mockReviews, Projects: projects}).Render(r.Context(), w)
	w.Write([]byte(`</div>`))
}

// HandleTaskDetail renders the task detail modal content
func (h *Handlers) HandleTaskDetailUI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	// Mock Task
	task := &views.UITask{
		ID:          id,
		ProjectID:   "web",
		Description: "Add Dark Mode to Dashboard",
		AgentName:   "Agent-042",
		Status:      "review",
		Progress:    100,
		MockDiff:    "file: src/App.css\n@@ -10,2 +10,3 @@\n :root {\n-  --bg: #fff;\n+  --bg: #111;\n+  --text: #eee;\n }",
		PreviewURL:  "https://cdn.dribbble.com/users/1615584/screenshots/15710288/media/7845f7478d59d56223253b8b603d1544.jpg?resize=400x300&vertical=center",
		Logs: []views.UILogEntry{
			{Type: "info", Message: "Task started"},
			{Type: "agent", Message: "Analyzing CSS files..."},
			{Type: "code", Message: "Generated dark mode variables"},
			{Type: "success", Message: "Tests passed"},
		},
	}
	
	project := views.UIProject{ID: "web", Name: "acme/web-dashboard", Icon: "fa-columns", Color: "text-purple-400"}

	components.TaskDetail(task, project).Render(r.Context(), w)
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

// HandleActionDiscard mocks discard action
func (h *Handlers) HandleActionDiscard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Task discarded"}`)
	h.HandleFeed(w, r)
}

// HandleActionChat mocks chat/refine action
func (h *Handlers) HandleActionChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("HX-Trigger", `{"close-modal": true, "toast": "Feedback sent"}`)
	h.HandleFeed(w, r)
}

// HandleAddTask mocks adding a task
func (h *Handlers) HandleAddTask(w http.ResponseWriter, r *http.Request) {
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
		UserID:       "default",
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
