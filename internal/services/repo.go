package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ErrMergeConflict indicates a merge conflict occurred.
type ErrMergeConflict struct {
	ConflictedFiles []string
	RepoPath        string
}

func (e *ErrMergeConflict) Error() string {
	return fmt.Sprintf("merge conflict in %d files: %s", len(e.ConflictedFiles), strings.Join(e.ConflictedFiles, ", "))
}

// GitManager handles git worktree operations.
type GitManager struct {
	repoRoot string
	dataDir  string
	mu       sync.Mutex
}

// NewGitManager creates a new repo manager.
// dataDir is the base directory for storing workspaces (e.g., "./data")
func NewGitManager(repoRoot, dataDir string) *GitManager {
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		absRoot = repoRoot
	}
	absDir, err := filepath.Abs(dataDir)
	if err != nil {
		absDir = dataDir // fallback if conversion fails
	}
	return &GitManager{repoRoot: absRoot, dataDir: absDir}
}

func (m *GitManager) Kind() RepoKind {
	return RepoKindGit
}

func (m *GitManager) RootPath() string {
	return m.repoRoot
}

// workspacePath returns the workspace path for a given task.
func (m *GitManager) workspacePath(taskID string) string {
	return filepath.Join(m.dataDir, "worktrees", "task-"+taskID)
}

// WorkspacePath returns the workspace path for a given task (exported).
func (m *GitManager) WorkspacePath(taskID string) string {
	return m.workspacePath(taskID)
}

// CreateWorkspace creates an isolated workspace for a task.
// Returns the workspace path.
func (m *GitManager) CreateWorkspace(ctx context.Context, taskID, branchName string) (string, error) {
	repoPath := m.repoRoot
	workspacePath := m.workspacePath(taskID)

	slog.Info("[GIT] Creating workspace", "task_id", taskID, "repo_path", repoPath, "workspace_path", workspacePath, "branch", branchName)

	// Check if workspace already exists
	if _, err := os.Stat(workspacePath); err == nil {
		slog.Info("[GIT] Workspace already exists", "path", workspacePath)
		return workspacePath, nil
	}

	// Ensure worktrees directory exists
	if err := os.MkdirAll(filepath.Dir(workspacePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktrees dir: %w", err)
	}

	slog.Info("[GIT] Executing: git worktree add -b", "branch", branchName, "path", workspacePath, "dir", repoPath)

	// Create new branch and workspace
	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branchName, workspacePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("[GIT] First attempt failed, trying without -b", "error", err, "output", string(output))
		// Branch might already exist, try without -b
		cmd = exec.CommandContext(ctx, "git", "worktree", "add", workspacePath, branchName)
		cmd.Dir = repoPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			slog.Error("[GIT] Worktree creation failed", "error", err, "output", string(output))
			return "", fmt.Errorf("git worktree add failed: %w\nOutput: %s", err, string(output))
		}
	}

	slog.Info("[GIT] Created workspace successfully", "task_id", taskID, "path", workspacePath, "branch", branchName)
	return workspacePath, nil
}

// Commit stages and commits changes without pushing.
func (m *GitManager) Commit(ctx context.Context, taskID, message string) error {
	workspacePath := m.workspacePath(taskID)

	slog.Info("[GIT] Commit called", "task_id", taskID, "workspace_path", workspacePath)

	// Stage all changes
	cmd := exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = workspacePath
	slog.Info("[GIT] Executing: git add -A", "dir", workspacePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Error("[GIT] git add failed", "error", err, "output", string(output))
		return fmt.Errorf("git add failed: %w\nOutput: %s", err, string(output))
	}
	slog.Info("[GIT] git add successful")

	// Check if there are changes to commit
	cmd = exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	cmd.Dir = workspacePath
	if err := cmd.Run(); err == nil {
		slog.Info("No changes to commit", "task_id", taskID)
		return nil
	}

	// Commit
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = workspacePath
	slog.Info("[GIT] Executing: git commit", "dir", workspacePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Error("[GIT] git commit failed", "error", err, "output", string(output))
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Committed successfully", "task_id", taskID)
	return nil
}

// CommitAndPush commits changes and pushes the branch.
func (m *GitManager) CommitAndPush(ctx context.Context, taskID, message string) error {
	if err := m.Commit(ctx, taskID, message); err != nil {
		return err
	}
	return m.PushBranch(ctx, taskID)
}

// PushBranch pushes the current branch to remote without committing.
func (m *GitManager) PushBranch(ctx context.Context, taskID string) error {
	workspacePath := m.workspacePath(taskID)

	cmd := exec.CommandContext(ctx, "git", "push", "-u", "origin", "HEAD")
	cmd.Dir = workspacePath
	slog.Info("[GIT] Executing: git push -u origin HEAD", "dir", workspacePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Error("[GIT] git push failed", "error", err, "output", string(output))
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Pushed branch successfully", "task_id", taskID)
	return nil
}

// GetCurrentBranch returns the current branch name in a workspace.
func (m *GitManager) GetCurrentBranch(ctx context.Context, taskID string) (string, error) {
	workspacePath := m.workspacePath(taskID)

	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = workspacePath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch failed: %w", err)
	}

	return string(output), nil
}

// GetDiff returns the git diff for a task's workspace.
// Shows diff between main branch and HEAD (all changes on the feature branch).
func (m *GitManager) GetDiff(ctx context.Context, taskID string) (string, error) {
	workspacePath := m.workspacePath(taskID)
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		slog.Warn("[GIT] Workspace missing, returning empty diff", "task_id", taskID, "path", workspacePath)
		return "", nil
	}

	// slog.Info("[GIT] GetDiff called", "task_id", taskID, "workspace_path", workspacePath)

	// Get current branch name
	branchCmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	branchCmd.Dir = workspacePath
	branchOutput, err := branchCmd.Output()
	if err != nil {
		slog.Error("[GIT] Failed to get branch name", "error", err)
		return "", fmt.Errorf("git branch failed: %w", err)
	}
	currentBranch := strings.TrimSpace(string(branchOutput))

	// Try origin/main first (remote tracking branch)
	cmd := exec.CommandContext(ctx, "git", "diff", "origin/main", currentBranch)
	cmd.Dir = workspacePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// slog.Warn("[GIT] GetDiff origin/main failed, trying main", "error", err)
		// Fallback to local main branch
		cmd = exec.CommandContext(ctx, "git", "diff", "main", currentBranch)
		cmd.Dir = workspacePath
		output, err = cmd.CombinedOutput()
		if err != nil {
			// slog.Warn("[GIT] GetDiff main failed, trying master", "error", err)
			// Try master branch
			cmd = exec.CommandContext(ctx, "git", "diff", "master", currentBranch)
			cmd.Dir = workspacePath
			output, err = cmd.CombinedOutput()
			if err != nil {
				slog.Error("[GIT] GetDiff failed for all branches", "error", err)
				return "", fmt.Errorf("git diff failed: %w\nOutput: %s", err, string(output))
			}
		}
	}

	slog.Info("[GIT] GetDiff successful", "task_id", taskID, "diff_size", len(output))
	return string(output), nil
}

// PullMainIntoWorktree pulls the latest main into the workspace and merges.
// If there's a merge conflict, returns ErrMergeConflict with the conflicted files.
func (m *GitManager) PullMainIntoWorktree(ctx context.Context, taskID string) error {
	workspacePath := m.workspacePath(taskID)
	repoPath := m.repoRoot

	slog.Info("[GIT] PullMainIntoWorktree called", "task_id", taskID, "workspace_path", workspacePath)

	// Fetch latest from origin in workspace
	cmd := exec.CommandContext(ctx, "git", "fetch", "origin", "main")
	cmd.Dir = workspacePath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try master
		cmd = exec.CommandContext(ctx, "git", "fetch", "origin", "master")
		cmd.Dir = workspacePath
		if output, err := cmd.CombinedOutput(); err != nil {
			slog.Warn("[GIT] Fetch failed", "error", err, "output", string(output))
		}
	}
	slog.Info("[GIT] Fetched latest from origin")

	// Also fetch in main repo to keep it updated
	cmd = exec.CommandContext(ctx, "git", "fetch", "origin")
	cmd.Dir = repoPath
	_ = cmd.Run()

	// Try to merge origin/main into the workspace
	cmd = exec.CommandContext(ctx, "git", "merge", "origin/main", "--no-edit")
	cmd.Dir = workspacePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a merge conflict
		if strings.Contains(string(output), "CONFLICT") || strings.Contains(string(output), "Automatic merge failed") {
			// Get list of conflicted files
			cmd = exec.CommandContext(ctx, "git", "diff", "--name-only", "--diff-filter=U")
			cmd.Dir = workspacePath
			conflictOutput, _ := cmd.Output()
			conflictedFiles := []string{}
			for _, f := range strings.Split(strings.TrimSpace(string(conflictOutput)), "\n") {
				if f != "" {
					conflictedFiles = append(conflictedFiles, f)
				}
			}
			slog.Info("[GIT] Merge conflict detected", "files", conflictedFiles)
			return &ErrMergeConflict{
				ConflictedFiles: conflictedFiles,
				RepoPath:        workspacePath,
			}
		}
		return fmt.Errorf("failed to merge origin/main: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Merged origin/main into workspace successfully")
	return nil
}

// CommitMergeResolution commits after merge conflict resolution.
func (m *GitManager) CommitMergeResolution(ctx context.Context, taskID, message string) error {
	workspacePath := m.workspacePath(taskID)

	// Stage all changes
	cmd := exec.CommandContext(ctx, "git", "add", "-A")
	cmd.Dir = workspacePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\nOutput: %s", err, string(output))
	}

	// Commit
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = workspacePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, string(output))
	}

	// Push
	cmd = exec.CommandContext(ctx, "git", "push", "-u", "origin", "HEAD")
	cmd.Dir = workspacePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Committed and pushed merge resolution", "task_id", taskID)
	return nil
}

// AbortMerge aborts an in-progress merge.
func (m *GitManager) AbortMerge(ctx context.Context, taskID string) error {
	workspacePath := m.workspacePath(taskID)

	cmd := exec.CommandContext(ctx, "git", "merge", "--abort")
	cmd.Dir = workspacePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git merge --abort failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Aborted merge", "task_id", taskID)
	return nil
}

// MergeToMain merges the task branch to main and pushes.
// Returns the branch name that was merged.
func (m *GitManager) MergeToMain(ctx context.Context, taskID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	repoPath := m.repoRoot
	workspacePath := m.workspacePath(taskID)

	slog.Info("[GIT] MergeToMain called", "task_id", taskID, "repo_path", repoPath, "workspace_path", workspacePath)

	// Get the branch name from the workspace
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = workspacePath
	branchOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get branch name: %w", err)
	}
	branchName := strings.TrimSpace(string(branchOutput))
	slog.Info("[GIT] Task branch", "branch", branchName)

	// Checkout main in the main repo
	cmd = exec.CommandContext(ctx, "git", "checkout", "main")
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try master
		cmd = exec.CommandContext(ctx, "git", "checkout", "master")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to checkout main/master: %w\nOutput: %s", err, string(output))
		}
	}
	slog.Info("[GIT] Checked out main branch")

	// Pull latest main
	cmd = exec.CommandContext(ctx, "git", "pull", "origin", "main")
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try master
		cmd = exec.CommandContext(ctx, "git", "pull", "origin", "master")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			slog.Warn("[GIT] Pull failed, continuing anyway", "error", err, "output", string(output))
		}
	}
	slog.Info("[GIT] Pulled latest main")

	// Merge the task branch
	cmd = exec.CommandContext(ctx, "git", "merge", branchName, "--no-edit")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check for merge conflict
		if strings.Contains(string(output), "CONFLICT") || strings.Contains(string(output), "Automatic merge failed") {
			// Abort the merge in main repo
			abortCmd := exec.CommandContext(ctx, "git", "merge", "--abort")
			abortCmd.Dir = repoPath
			_ = abortCmd.Run()

			// Get list of conflicted files
			conflictedFiles := []string{}
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "CONFLICT") && strings.Contains(line, "Merge conflict in") {
					parts := strings.Split(line, "Merge conflict in ")
					if len(parts) > 1 {
						conflictedFiles = append(conflictedFiles, strings.TrimSpace(parts[1]))
					}
				}
			}
			slog.Info("[GIT] Merge conflict detected in MergeToMain", "files", conflictedFiles)
			return "", &ErrMergeConflict{
				ConflictedFiles: conflictedFiles,
				RepoPath:        workspacePath,
			}
		}
		return "", fmt.Errorf("failed to merge branch %s: %w\nOutput: %s", branchName, err, string(output))
	}
	slog.Info("[GIT] Merged branch", "branch", branchName)

	// Push to origin
	cmd = exec.CommandContext(ctx, "git", "push", "origin", "main")
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try master
		cmd = exec.CommandContext(ctx, "git", "push", "origin", "master")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to push to main: %w\nOutput: %s", err, string(output))
		}
	}
	slog.Info("[GIT] Pushed to origin main")

	// Delete the remote branch (optional, don't fail if this errors)
	cmd = exec.CommandContext(ctx, "git", "push", "origin", "--delete", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("[GIT] Failed to delete remote branch (may not exist)", "branch", branchName, "error", err, "output", string(output))
	} else {
		slog.Info("[GIT] Deleted remote branch", "branch", branchName)
	}

	// Remove the workspace first (branch cannot be deleted while used by workspace)
	if err := os.RemoveAll(workspacePath); err != nil {
		slog.Warn("[GIT] Failed to remove workspace directory", "error", err)
	} else {
		slog.Info("[GIT] Removed workspace directory", "path", workspacePath)
	}

	// Prune workspaces to clean up git's internal state
	cmd = exec.CommandContext(ctx, "git", "worktree", "prune")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("[GIT] Failed to prune workspaces", "error", err, "output", string(output))
	} else {
		slog.Info("[GIT] Pruned workspaces")
	}

	// Delete the local branch
	cmd = exec.CommandContext(ctx, "git", "branch", "-d", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("[GIT] Failed to delete local branch", "branch", branchName, "error", err, "output", string(output))
	} else {
		slog.Info("[GIT] Deleted local branch", "branch", branchName)
	}

	slog.Info("[GIT] MergeToMain completed successfully", "task_id", taskID, "branch", branchName)
	return branchName, nil
}

// RemoveWorkspace removes the workspace for a task.
func (m *GitManager) RemoveWorkspace(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	workspacePath := m.workspacePath(taskID)
	slog.Info("[GIT] Removing workspace", "task_id", taskID, "path", workspacePath)

	// Check if workspace exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		slog.Info("[GIT] Workspace does not exist, nothing to remove", "task_id", taskID)
		return nil
	}

	// Remove the workspace directory first
	if err := os.RemoveAll(workspacePath); err != nil {
		slog.Error("[GIT] Failed to remove workspace directory", "error", err)
		return fmt.Errorf("failed to remove workspace directory: %w", err)
	}

	// Prune workspaces to clean up git's internal state
	cmd := exec.CommandContext(ctx, "git", "worktree", "prune")
	cmd.Dir = m.repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("[GIT] Failed to prune workspaces", "error", err, "output", string(output))
	}

	slog.Info("[GIT] Workspace removed", "task_id", taskID)
	return nil
}
