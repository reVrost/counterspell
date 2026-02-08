package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
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

func (h *Handlers) HandleListWorkspaceFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := chi.URLParam(r, "id")
	if workspaceID == "" {
		_ = render.Render(w, r, ErrInvalidRequest(errors.New("workspace ID required")))
		return
	}

	path := r.URL.Query().Get("path")
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	limit := 200
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	workspace, err := h.repository.GetWorkspace(ctx, workspaceID)
	if err != nil {
		slog.Error("Failed to get workspace", "workspace_id", workspaceID, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to get workspace", err))
		return
	}

	files, err := h.fileService.ListWorkspaceFiles(ctx, workspace.LocalPath, path, limit, filter)
	if err != nil {
		slog.Error("Failed to list workspace files", "workspace_id", workspaceID, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to list workspace files", err))
		return
	}

	render.JSON(w, r, files)
}

func (h *Handlers) HandleUploadWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := chi.URLParam(r, "id")
	if workspaceID == "" {
		_ = render.Render(w, r, ErrInvalidRequest(errors.New("workspace ID required")))
		return
	}

	if r.Method != "POST" {
		_ = render.Render(w, r, ErrInvalidRequest(errors.New("method not allowed")))
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("failed to parse multipart form: %w", err)))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("failed to get file from form: %w", err)))
		return
	}
	defer file.Close()

	const maxFileSize = 25 * 1024 * 1024
	if header.Size > maxFileSize {
		_ = render.Render(w, r, ErrInvalidRequest(fmt.Errorf("file size exceeds 25MB limit")))
		return
	}

	workspace, err := h.repository.GetWorkspace(ctx, workspaceID)
	if err != nil {
		slog.Error("Failed to get workspace", "workspace_id", workspaceID, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to get workspace", err))
		return
	}

	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = filepath.Join("uploads", header.Filename)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		_ = render.Render(w, r, ErrInternalServer("Failed to read file", err))
		return
	}

	savedPath, err := h.fileService.UploadWorkspaceFile(ctx, workspace.LocalPath, targetPath, content)
	if err != nil {
		slog.Error("Failed to upload workspace file", "workspace_id", workspaceID, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to upload file", err))
		return
	}

	render.JSON(w, r, map[string]string{"path": savedPath})
}

func (h *Handlers) HandleReadWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := chi.URLParam(r, "id")
	if workspaceID == "" {
		_ = render.Render(w, r, ErrInvalidRequest(errors.New("workspace ID required")))
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		_ = render.Render(w, r, ErrInvalidRequest(errors.New("path parameter required")))
		return
	}

	workspace, err := h.repository.GetWorkspace(ctx, workspaceID)
	if err != nil {
		slog.Error("Failed to get workspace", "workspace_id", workspaceID, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to get workspace", err))
		return
	}

	fileInfo, preview, err := h.fileService.ReadWorkspaceFile(ctx, workspace.LocalPath, path)
	if err != nil {
		slog.Error("Failed to read workspace file", "workspace_id", workspaceID, "path", path, "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to read file", err))
		return
	}

	render.JSON(w, r, map[string]any{
		"file_info": fileInfo,
		"preview":   preview,
	})
}
