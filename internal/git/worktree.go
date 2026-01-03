package git

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

// WorktreeManager manages git worktrees for isolated agent workspaces.
type WorktreeManager struct {
	repoPath string
}

// NewWorktreeManager creates a new worktree manager.
func NewWorktreeManager(repoPath string) *WorktreeManager {
	return &WorktreeManager{repoPath: repoPath}
}

// CreateWorktree creates a new worktree for a task.
func (m *WorktreeManager) CreateWorktree(taskID string, baseBranch string) (string, error) {
	fullPath := filepath.Join(m.repoPath, "..", "worktree-"+taskID)

	// Check if worktree already exists
	if _, err := os.Stat(fullPath); err == nil {
		slog.Info("Worktree already exists", "path", fullPath)
		return fullPath, nil
	}

	// Create worktree
	cmd := exec.Command("git", "worktree", "add", fullPath, baseBranch)
	cmd.Dir = m.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create worktree: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Created worktree", "task_id", taskID, "path", fullPath)
	return fullPath, nil
}

// CleanupWorktree removes a worktree after task completion.
func (m *WorktreeManager) CleanupWorktree(taskID string) error {
	fullPath := filepath.Join(m.repoPath, "..", "worktree-"+taskID)

	// Remove worktree
	cmd := exec.Command("git", "worktree", "remove", fullPath)
	cmd.Dir = m.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Removed worktree", "task_id", taskID)
	return nil
}

// PruneWorktrees removes all stale worktrees.
func (m *WorktreeManager) PruneWorktrees() error {
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = m.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to prune worktrees: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Pruned worktrees")
	return nil
}
