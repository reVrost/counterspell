package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// EnsureDirectories creates the required directory structure for multi-tenant mode.
// Directory structure:
//
//	data/
//	├── db/           # Per-user SQLite databases
//	├── repos/        # Shared bare git repos
//	└── workspaces/   # User-isolated worktrees
func EnsureDirectories(cfg *Config) error {
	dirs := []string{
		filepath.Join(cfg.DataDir, "db"),
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

// MigrateFromSingleUser migrates existing single-user data to multi-tenant structure.
// This is called on startup when MULTI_TENANT=false to ensure backward compatibility.
func MigrateFromSingleUser(cfg *Config) error {
	// Old paths (pre-multi-tenant)
	oldDBPath := filepath.Join(cfg.DataDir, "counterspell.db")
	oldWorktreesPath := filepath.Join(cfg.DataDir, "worktrees")

	// New paths
	newDBPath := cfg.DBPath("default")
	newWorkspacePath := cfg.WorkspacesPath("default")
	newWorktreesPath := filepath.Join(newWorkspacePath, "worktrees")

	// Migrate database file if exists
	if _, err := os.Stat(oldDBPath); err == nil {
		// Only migrate if new path doesn't exist
		if _, err := os.Stat(newDBPath); os.IsNotExist(err) {
			slog.Info("Migrating database to new location", "from", oldDBPath, "to", newDBPath)
			if err := os.MkdirAll(filepath.Dir(newDBPath), 0755); err != nil {
				return fmt.Errorf("failed to create db directory: %w", err)
			}
			if err := os.Rename(oldDBPath, newDBPath); err != nil {
				return fmt.Errorf("failed to migrate database: %w", err)
			}
		}
	}

	// Migrate worktrees (move task-* directories to user workspace)
	if _, err := os.Stat(oldWorktreesPath); err == nil {
		entries, err := os.ReadDir(oldWorktreesPath)
		if err != nil {
			return fmt.Errorf("failed to read old worktrees: %w", err)
		}

		if err := os.MkdirAll(newWorktreesPath, 0755); err != nil {
			return fmt.Errorf("failed to create new worktrees directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			oldPath := filepath.Join(oldWorktreesPath, entry.Name())
			newPath := filepath.Join(newWorktreesPath, entry.Name())

			// Only migrate if destination doesn't exist
			if _, err := os.Stat(newPath); os.IsNotExist(err) {
				slog.Info("Migrating worktree", "from", oldPath, "to", newPath)
				if err := os.Rename(oldPath, newPath); err != nil {
					slog.Warn("Failed to migrate worktree", "path", oldPath, "error", err)
				}
			}
		}

		// Remove old worktrees directory if empty
		if isEmpty, _ := isDirEmpty(oldWorktreesPath); isEmpty {
			_ = os.Remove(oldWorktreesPath)
		}
	}

	// Note: repos stay in place (data/repos/{owner}/{repo}) - no migration needed
	// But we might need to convert them to bare repos in the future

	return nil
}

func isDirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
