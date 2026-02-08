package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// FileService handles file operations for agents.
type FileService struct {
	dataDir string
}

// NewFileService creates a new file service.
func NewFileService(dataDir string) *FileService {
	return &FileService{dataDir: dataDir}
}

// FileInfo represents file metadata.
type FileInfo struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	IsDir     bool      `json:"is_dir"`
	Extension string    `json:"extension,omitempty"`
}

// Search searches for files matching a pattern.
func (s *FileService) Search(ctx context.Context, pattern, directory string, maxResults int) ([]FileInfo, error) {
	slog.Info("[FILE] Searching files", "pattern", pattern, "directory", directory)

	// Default to data dir if no directory specified
	if directory == "" {
		directory = s.dataDir
	}

	var results []FileInfo

	// Walk directory tree
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		// Skip .git, node_modules, etc.
		relPath, _ := filepath.Rel(directory, path)
		for _, skip := range []string{".git", "node_modules", ".next", ".cache", "target", "build"} {
			if strings.Contains(relPath, skip) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check if matches pattern
		if pattern != "" {
			if !strings.Contains(strings.ToLower(filepath.Base(path)), strings.ToLower(pattern)) {
				return nil
			}
		}

		// Skip directories in results (unless specifically looking for dirs)
		if info.IsDir() {
			return nil
		}

		// Add to results
		results = append(results, FileInfo{
			Path:      relPath,
			Name:      info.Name(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			IsDir:     info.IsDir(),
			Extension: filepath.Ext(path),
		})

		// Limit results
		if maxResults > 0 && len(results) >= maxResults {
			return fmt.Errorf("max results reached")
		}

		return nil
	})

	// If we hit max results, it's not an error
	if err != nil && err.Error() != "max results reached" {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	slog.Info("[FILE] Search complete", "results", len(results))
	return results, nil
}

// Read reads file contents.
func (s *FileService) Read(ctx context.Context, path string) (string, error) {
	slog.Info("[FILE] Reading file", "path", path)

	// Resolve path relative to data dir
	fullPath := filepath.Join(s.dataDir, path)

	// Security check: ensure path is within data dir
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	absDataDir, err := filepath.Abs(s.dataDir)
	if err != nil {
		return "", fmt.Errorf("invalid data dir: %w", err)
	}

	if !strings.HasPrefix(absPath, absDataDir) {
		return "", fmt.Errorf("path outside data directory")
	}

	// Read file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	slog.Info("[FILE] File read", "path", path, "size", len(content))
	return string(content), nil
}

// Write writes content to a file.
func (s *FileService) Write(ctx context.Context, path, content string) error {
	slog.Info("[FILE] Writing file", "path", path, "size", len(content))

	// Resolve path relative to data dir
	fullPath := filepath.Join(s.dataDir, path)

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	slog.Info("[FILE] File written", "path", path)
	return nil
}

// Delete deletes a file.
func (s *FileService) Delete(ctx context.Context, path string) error {
	slog.Info("[FILE] Deleting file", "path", path)

	// Resolve path relative to data dir
	fullPath := filepath.Join(s.dataDir, path)

	// Delete file
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	slog.Info("[FILE] File deleted", "path", path)
	return nil
}

// List lists files in a directory.
func (s *FileService) List(ctx context.Context, directory string) ([]FileInfo, error) {
	slog.Info("[FILE] Listing directory", "directory", directory)

	// Default to data dir if no directory specified
	if directory == "" {
		directory = s.dataDir
	}

	fullPath := filepath.Join(s.dataDir, directory)

	// Read directory
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var results []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		results = append(results, FileInfo{
			Path:      entry.Name(),
			Name:      entry.Name(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			IsDir:     entry.IsDir(),
			Extension: filepath.Ext(entry.Name()),
		})
	}

	slog.Info("[FILE] Directory listed", "directory", directory, "count", len(results))
	return results, nil
}

// GetProjectRoot finds the project root directory.
func (s *FileService) GetProjectRoot(ctx context.Context) (string, error) {
	// Start from current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Look for project indicators (.git, package.json, go.mod, etc.)
	indicators := []string{".git", "package.json", "go.mod", "Cargo.toml", "pyproject.toml"}

	// Walk up directory tree
	dir := cwd
	for {
		for _, indicator := range indicators {
			if _, err := os.Stat(filepath.Join(dir, indicator)); err == nil {
				slog.Info("[FILE] Found project root", "root", dir)
				return dir, nil
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding indicators
			break
		}
		dir = parent
	}

	// Fallback to current directory
	slog.Info("[FILE] No project indicators found, using cwd", "cwd", cwd)
	return cwd, nil
}

// GetPlatformInfo returns platform information.
func (s *FileService) GetPlatformInfo() map[string]any {
	return map[string]any{
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"separator": string(filepath.Separator),
	}
}

// PathJoin joins path components safely.
func (s *FileService) PathJoin(parts ...string) string {
	return filepath.Join(parts...)
}

// ListWorkspaceFiles lists files in a workspace directory with filtering and limits.
func (s *FileService) ListWorkspaceFiles(ctx context.Context, workspacePath, relativePath string, limit int, filter string) ([]FileInfo, error) {
	slog.Info("[FILE] Listing workspace files", "workspace_path", workspacePath, "relative_path", relativePath, "filter", filter, "limit", limit)

	var results []FileInfo
	fullPath := filepath.Join(workspacePath, relativePath)

	if relativePath != "" {
		if err := s.validatePath(workspacePath, fullPath); err != nil {
			return nil, err
		}
	}

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		relPath, err := filepath.Rel(workspacePath, path)
		if err != nil {
			return nil
		}

		if path == fullPath && info.IsDir() {
			return nil
		}

		skipDirs := []string{".git", "node_modules", ".next", ".cache", "target", "build", "dist", ".venv", "venv", "__pycache__", ".idea", ".vscode", ".DS_Store", "Thumbs.db"}
		for _, skip := range skipDirs {
			if strings.Contains(relPath, skip) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if filter != "" {
			switch filter {
			case "recent":
				if time.Since(info.ModTime()) > 7*24*time.Hour {
					return nil
				}
			case "images":
				ext := strings.ToLower(filepath.Ext(path))
				imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".ico", ".bmp"}
				isImage := false
				for _, ie := range imageExts {
					if ext == ie {
						isImage = true
						break
					}
				}
				if !isImage {
					return nil
				}
			case "docs":
				ext := strings.ToLower(filepath.Ext(path))
				docExts := []string{".txt", ".md", ".pdf", ".doc", ".docx", ".rtf", ".odt"}
				isDoc := false
				for _, de := range docExts {
					if ext == de {
						isDoc = true
						break
					}
				}
				if !isDoc {
					return nil
				}
			}
		}

		results = append(results, FileInfo{
			Path:      relPath,
			Name:      info.Name(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			IsDir:     info.IsDir(),
			Extension: filepath.Ext(path),
		})

		if limit > 0 && len(results) >= limit {
			return fmt.Errorf("max results reached")
		}

		return nil
	})

	if err != nil && err.Error() != "max results reached" {
		return nil, fmt.Errorf("failed to list workspace files: %w", err)
	}

	slog.Info("[FILE] Workspace files listed", "count", len(results))
	return results, nil
}

// UploadWorkspaceFile uploads a file to a workspace.
func (s *FileService) UploadWorkspaceFile(ctx context.Context, workspacePath, relativePath string, content []byte) (string, error) {
	slog.Info("[FILE] Uploading workspace file", "workspace_path", workspacePath, "relative_path", relativePath, "size", len(content))

	if relativePath == "" {
		relativePath = "uploads/" + filepath.Base(relativePath)
	}

	fullPath := filepath.Join(workspacePath, relativePath)

	if err := s.validatePath(workspacePath, fullPath); err != nil {
		return "", err
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	slog.Info("[FILE] Workspace file uploaded", "path", relativePath)
	return relativePath, nil
}

// ReadWorkspaceFile reads a file from a workspace with preview support.
func (s *FileService) ReadWorkspaceFile(ctx context.Context, workspacePath, relativePath string) (FileInfo, string, error) {
	slog.Info("[FILE] Reading workspace file", "workspace_path", workspacePath, "relative_path", relativePath)

	fullPath := filepath.Join(workspacePath, relativePath)

	if err := s.validatePath(workspacePath, fullPath); err != nil {
		return FileInfo{}, "", err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return FileInfo{}, "", fmt.Errorf("failed to stat file: %w", err)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return FileInfo{}, "", fmt.Errorf("failed to read file: %w", err)
	}

	fileInfo := FileInfo{
		Path:      relativePath,
		Name:      info.Name(),
		Size:      info.Size(),
		ModTime:   info.ModTime(),
		IsDir:     info.IsDir(),
		Extension: filepath.Ext(relativePath),
	}

	binaryExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".zip", ".tar", ".gz", ".exe", ".dll", ".so", ".dylib", ".mp3", ".mp4", ".avi", ".mov", ".woff", ".woff2", ".ttf", ".eot"}
	isBinary := false
	for _, ext := range binaryExts {
		if strings.EqualFold(fileInfo.Extension, ext) {
			isBinary = true
			break
		}
	}

	var preview string
	if !isBinary && info.Size() < 1024*1024 {
		contentStr := string(content)
		if isText(contentStr) {
			if len(contentStr) > 10000 {
				preview = contentStr[:10000] + "\n... (truncated)"
			} else {
				preview = contentStr
			}
		} else {
			preview = ""
		}
	}

	slog.Info("[FILE] Workspace file read", "path", relativePath, "size", info.Size(), "has_preview", preview != "")
	return fileInfo, preview, nil
}

// SearchWorkspaceFiles searches for files in a workspace.
func (s *FileService) SearchWorkspaceFiles(ctx context.Context, workspacePath, query string, maxResults int) ([]string, error) {
	slog.Info("[FILE] Searching workspace files", "workspace_path", workspacePath, "query", query, "max_results", maxResults)

	var results []string
	lowerQuery := strings.ToLower(query)

	err := filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		relPath, err := filepath.Rel(workspacePath, path)
		if err != nil {
			return nil
		}

		skipDirs := []string{".git", "node_modules", ".next", ".cache", "target", "build", "dist", ".venv", "venv", "__pycache__", ".idea", ".vscode", ".DS_Store"}
		for _, skip := range skipDirs {
			if strings.Contains(relPath, skip) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}

		if strings.Contains(strings.ToLower(relPath), lowerQuery) {
			results = append(results, relPath)
		}

		if maxResults > 0 && len(results) >= maxResults {
			return fmt.Errorf("max results reached")
		}

		return nil
	})

	if err != nil && err.Error() != "max results reached" {
		return nil, fmt.Errorf("failed to search workspace files: %w", err)
	}

	slog.Info("[FILE] Workspace file search complete", "results", len(results))
	return results, nil
}

// validatePath ensures a path is within the workspace directory.
func (s *FileService) validatePath(workspacePath, targetPath string) error {
	absWorkspace, err := filepath.Abs(workspacePath)
	if err != nil {
		return fmt.Errorf("invalid workspace path: %w", err)
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("invalid target path: %w", err)
	}

	if !strings.HasPrefix(absTarget, absWorkspace) {
		return fmt.Errorf("path outside workspace directory")
	}

	return nil
}

// isText checks if content appears to be text.
func isText(content string) bool {
	if len(content) == 0 {
		return true
	}

	nullCount := 0
	for i := 0; i < len(content) && i < 1024; i++ {
		if content[i] == 0 {
			nullCount++
		}
		if nullCount > 1 {
			return false
		}
	}

	for _, r := range content {
		if r == 0 {
			return false
		}
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}

	return true
}
