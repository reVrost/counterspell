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
