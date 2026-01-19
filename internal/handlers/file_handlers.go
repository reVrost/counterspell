package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

// HandleFileRead reads a file.
func (h *Handlers) HandleFileRead(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	content, err := h.fileService.Read(ctx, path)
	if err != nil {
		slog.Error("Failed to read file", "error", err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]string{"content": content, "path": path})
}

// HandleFileWrite writes to a file.
func (h *Handlers) HandleFileWrite(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.fileService.Write(ctx, path, req.Content); err != nil {
		slog.Error("Failed to write file", "error", err)
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}

// HandleFileList lists files in a directory.
func (h *Handlers) HandleFileList(w http.ResponseWriter, r *http.Request) {
	directory := r.URL.Query().Get("directory")

	ctx := r.Context()
	files, err := h.fileService.List(ctx, directory)
	if err != nil {
		slog.Error("Failed to list files", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to list files", err))
		return
	}

	render.JSON(w, r, files)
}

// HandleFileDelete deletes a file.
func (h *Handlers) HandleFileDelete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.fileService.Delete(ctx, path); err != nil {
		slog.Error("Failed to delete file", "error", err)
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]string{"status": "ok"})
}

// HandleProjectRoot finds project root.
func (h *Handlers) HandleProjectRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	root, err := h.fileService.GetProjectRoot(ctx)
	if err != nil {
		slog.Error("Failed to find project root", "error", err)
		http.Error(w, "Failed to find project root", http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]string{"root": root})
}

// HandlePlatformInfo returns platform information.
func (h *Handlers) HandlePlatformInfo(w http.ResponseWriter, r *http.Request) {
	info := h.fileService.GetPlatformInfo()
	render.JSON(w, r, info)
}

// HandleToolExecution executes a tool operation.
func (h *Handlers) HandleToolExecution(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Tool  string         `json:"tool"`
		Input map[string]any `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Execute tool
	ctx := r.Context()
	result, err := h.toolService.ExecuteTool(ctx, req.Tool, req.Input)
	if err != nil {
		slog.Error("Failed to execute tool", "tool", req.Tool, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]any{
		"tool":   req.Tool,
		"result": result,
	})
}
