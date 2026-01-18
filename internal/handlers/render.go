package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/revrost/code/counterspell/internal/services"
)

// ------------------------------------------------------------------
// Request types (implement render.Binder)
// ------------------------------------------------------------------

// AddTaskRequest is the request payload for adding a task.
type AddTaskRequest struct {
	VoiceInput string `json:"voice_input"`
	ProjectID  string `json:"project_id"`
	ModelID    string `json:"model_id"`
}

func (a *AddTaskRequest) Bind(r *http.Request) error {
	if a.VoiceInput == "" {
		return errors.New("voice_input is required")
	}
	if a.ProjectID == "" {
		return errors.New("project_id is required")
	}
	return nil
}

// ChatRequest is the request payload for continuing a task.
type ChatRequest struct {
	Message string `json:"message"`
	ModelID string `json:"model_id"`
}

func (c *ChatRequest) Bind(r *http.Request) error {
	if c.Message == "" {
		return errors.New("message is required")
	}
	if c.ModelID == "" {
		c.ModelID = "o#anthropic/claude-sonnet-4"
	}
	return nil
}

// ResolveConflictRequest is the request payload for resolving merge conflicts.
type ResolveConflictRequest struct {
	File   string `json:"file"`
	Choice string `json:"choice"` // "ours" or "theirs"
}

func (rc *ResolveConflictRequest) Bind(r *http.Request) error {
	if rc.File == "" {
		return errors.New("file is required")
	}
	if rc.Choice != "ours" && rc.Choice != "theirs" {
		return errors.New("choice must be 'ours' or 'theirs'")
	}
	return nil
}

// ActivateProjectRequest is the request payload for activating a project.
type ActivateProjectRequest struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

func (ap *ActivateProjectRequest) Bind(r *http.Request) error {
	if ap.Owner == "" || ap.Repo == "" {
		return errors.New("owner and repo are required")
	}
	return nil
}

// SaveSettingsRequest is the request payload for saving user settings.
type SaveSettingsRequest struct {
	OpenRouterKey string `json:"openrouter_key"`
	ZaiKey        string `json:"zai_key"`
	AnthropicKey  string `json:"anthropic_key"`
	OpenAIKey     string `json:"openai_key"`
	AgentBackend  string `json:"agent_backend"`
}

func (s *SaveSettingsRequest) Bind(r *http.Request) error {
	return nil
}

// ------------------------------------------------------------------
// Response types (implement render.Renderer)
// ------------------------------------------------------------------

// StatusResponse is a generic success/error response.
type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	PRURL   string `json:"pr_url,omitempty"`
}

func (sr *StatusResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// ConflictResponse is returned when a merge has conflicts.
type ConflictResponse struct {
	Status    string                  `json:"status"`
	TaskID    string                  `json:"task_id"`
	Conflicts []services.ConflictFile `json:"conflicts"`
}

func (cr *ConflictResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// ------------------------------------------------------------------
// Error responses
// ------------------------------------------------------------------

// ErrResponse is a structured error response.
type ErrResponse struct {
	Err            error  `json:"-"`
	HTTPStatusCode int    `json:"-"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// ErrInvalidRequest returns a 400 Bad Request error.
func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		Status:         "error",
		Message:        err.Error(),
	}
}

// ErrNotFound returns a 404 Not Found error.
func ErrNotFound(msg string) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusNotFound,
		Status:         "error",
		Message:        msg,
	}
}

// ErrInternalServer returns a 500 Internal Server Error.
func ErrInternalServer(msg string) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusInternalServerError,
		Status:         "error",
		Message:        msg,
	}
}

// ErrUnauthorized returns a 401 Unauthorized error.
func ErrUnauthorized(msg string) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusUnauthorized,
		Status:         "error",
		Message:        msg,
	}
}

// ------------------------------------------------------------------
// Helper constructors
// ------------------------------------------------------------------

// Success returns a success status response.
func Success(message string) render.Renderer {
	return &StatusResponse{Status: "success", Message: message}
}

// SuccessWithPR returns a success response with PR URL.
func SuccessWithPR(message, prURL string) render.Renderer {
	return &StatusResponse{Status: "success", Message: message, PRURL: prURL}
}

// Conflict returns a conflict response for merge conflicts.
func Conflict(taskID string, conflicts []services.ConflictFile) render.Renderer {
	return &ConflictResponse{
		Status:    "conflict",
		TaskID:    taskID,
		Conflicts: conflicts,
	}
}
