package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/render"
	"github.com/revrost/counterspell/internal/db/sqlc"
)

const (
	workspaceIcon  = "fa-folder"
	workspaceColor = "text-violet-400"
)

func (h *Handlers) HandleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaces, err := h.repository.ListWorkspaces(ctx)
	if err != nil {
		slog.Error("Failed to list workspaces", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to load workspaces", err))
		return
	}

	response := make([]ProjectResponse, 0, len(workspaces))
	for _, workspace := range workspaces {
		response = append(response, workspaceToProjectResponse(workspace))
	}

	render.JSON(w, r, response)
}

func (h *Handlers) HandleWorkspaceSetup(w http.ResponseWriter, r *http.Request) {
	cwdPath, err := os.Getwd()
	if err != nil {
		slog.Error("Failed to resolve current directory", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to resolve current directory", err))
		return
	}

	cwdName := filepath.Base(cwdPath)
	if cwdName == "." || cwdName == string(filepath.Separator) {
		cwdName = "workspace"
	}

	render.JSON(w, r, map[string]string{
		"cwd_path": cwdPath,
		"cwd_name": cwdName,
	})
}

func (h *Handlers) HandleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Name string `json:"name"`
		Mode string `json:"mode"`
	}

	if err := render.DecodeJSON(r.Body, &req); err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("invalid request: %w", err)))
		return
	}

	mode := strings.TrimSpace(req.Mode)
	if mode != "import_current" && mode != "new_folder" {
		_ = render.Render(w, r, ErrInvalidRequest(errors.New("mode must be import_current or new_folder")))
		return
	}

	cwdPath, err := os.Getwd()
	if err != nil {
		slog.Error("Failed to resolve current directory", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to resolve current directory", err))
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = filepath.Base(cwdPath)
	}
	if name == "" || name == "." || name == string(filepath.Separator) {
		_ = render.Render(w, r, ErrInvalidRequest(errors.New("workspace name is required")))
		return
	}

	var localPath string
	switch mode {
	case "import_current":
		localPath = cwdPath
	case "new_folder":
		localPath = filepath.Join(cwdPath, name)
		if err := os.MkdirAll(localPath, 0o755); err != nil {
			_ = render.Render(w, r, ErrInternalServer("Failed to create workspace folder", err))
			return
		}
	}

	workspace, err := h.repository.CreateWorkspace(ctx, name, localPath)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "UNIQUE constraint failed: workspaces.name") {
			_ = render.Render(w, r, ErrInvalidRequest(errors.New("workspace name already exists")))
			return
		}
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(errMsg, "failed to load github connection") {
			_ = render.Render(w, r, ErrInvalidRequest(errors.New("github connection is required before creating a workspace")))
			return
		}

		slog.Error("Failed to create workspace", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to create workspace", err))
		return
	}

	response := workspaceToProjectResponse(workspace)
	render.JSON(w, r, map[string]any{
		"id":         response.ID,
		"name":       response.Name,
		"icon":       response.Icon,
		"color":      response.Color,
		"local_path": workspace.LocalPath,
	})
}

func workspaceToProjectResponse(workspace sqlc.Workspace) ProjectResponse {
	return ProjectResponse{
		ID:    workspace.ID,
		Name:  workspace.Name,
		Icon:  workspaceIcon,
		Color: workspaceColor,
	}
}
