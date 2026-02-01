package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// EnsureDirectories creates the required directory structure.
// Directory structure:
//
//	data/
//	├── repos/        # Shared bare git repos
//	└── workspaces/   # User-isolated worktrees
func EnsureDirectories(cfg *Config) error {
	dirs := []string{
		filepath.Join(cfg.DataDir, "repos"),
		filepath.Join(cfg.DataDir, "workspaces"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		slog.Debug("Ensured directory exists", "path", dir)
	}

	slog.Info("Directory structure ready", "data_dir", cfg.DataDir)
	return nil
}

// EnsureUserWorkspace creates the workspace directory for a specific user.
func EnsureUserWorkspace(cfg *Config, userID string) error {
	workspacePath := cfg.WorkspacesPath(userID)
	worktreesPath := filepath.Join(workspacePath, "worktrees")

	if err := os.MkdirAll(worktreesPath, 0755); err != nil {
		return fmt.Errorf("failed to create user workspace: %w", err)
	}

	return nil
}
