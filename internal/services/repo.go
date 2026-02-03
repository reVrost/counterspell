package services

import (
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

// GitManager handles git workspace operations.
type GitManager struct {
	repoRoot string
	dataDir  string
	mu       sync.Mutex
}

// NewGitManager creates a new repo manager.
// repoRoot is the root of the local git repository.
// dataDir is the base directory for storing workspaces (e.g., "./data").
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
// Returns the worktree path.
func (m *GitManager) CreateWorktree(owner, repo, taskID, branchName string) (string, error) {
	repoPath := m.repoPath(owner, repo)
	worktreePath := m.worktreePath(taskID)

	slog.Info("[GIT] Creating worktree", "task_id", taskID, "repo_path", repoPath, "worktree_path", worktreePath, "branch", branchName)

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		slog.Info("[GIT] Worktree already exists", "path", worktreePath)
		return worktreePath, nil
	}

	// Ensure worktrees directory exists
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktrees dir: %w", err)
	}

	slog.Info("[GIT] Executing: git worktree add -b", "branch", branchName, "path", worktreePath, "dir", repoPath)

	// Create new branch and worktree
	cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("[GIT] First attempt failed, trying without -b", "error", err, "output", string(output))
		// Branch might already exist, try without -b
		cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
		cmd.Dir = repoPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			slog.Error("[GIT] Worktree creation failed", "error", err, "output", string(output))
			return "", fmt.Errorf("git worktree add failed: %w\nOutput: %s", err, string(output))
		}
	}

	slog.Info("[GIT] Created worktree successfully", "task_id", taskID, "path", worktreePath, "branch", branchName)
	return worktreePath, nil
}

// CleanupWorktree removes a worktree.
func (m *GitManager) CleanupWorktree(owner, repo, taskID string) error {
	repoPath := m.repoPath(owner, repo)
	worktreePath := m.worktreePath(taskID)

	cmd := exec.Command("git", "worktree", "remove", "--force", worktreePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Removed worktree", "task_id", taskID)
	return nil
}

// Commit stages and commits changes without pushing.
func (m *GitManager) Commit(taskID, message string) error {
	worktreePath := m.worktreePath(taskID)

	slog.Info("[GIT] Commit called", "task_id", taskID, "worktree_path", worktreePath)

	// Stage all changes
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = worktreePath
	slog.Info("[GIT] Executing: git add -A", "dir", worktreePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Error("[GIT] git add failed", "error", err, "output", string(output))
		return fmt.Errorf("git add failed: %w\nOutput: %s", err, string(output))
	}
	slog.Info("[GIT] git add successful")

	// Check if there are changes to commit
	cmd = exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = worktreePath
	if err := cmd.Run(); err == nil {
		slog.Info("No changes to commit", "task_id", taskID)
		return nil
	}

	// Commit
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = worktreePath
	slog.Info("[GIT] Executing: git commit", "dir", worktreePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Error("[GIT] git commit failed", "error", err, "output", string(output))
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Committed successfully", "task_id", taskID)
	return nil
}

// CommitAndPush commits changes and pushes the branch.
func (m *GitManager) CommitAndPush(taskID, message string) error {
	if err := m.Commit(taskID, message); err != nil {
		return err
	}
	return m.PushBranch(taskID)
}

// PushBranch pushes the current branch to remote without committing.
func (m *GitManager) PushBranch(taskID string) error {
	worktreePath := m.worktreePath(taskID)

	cmd := exec.Command("git", "push", "-u", "origin", "HEAD")
	cmd.Dir = worktreePath
	slog.Info("[GIT] Executing: git push -u origin HEAD", "dir", worktreePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Error("[GIT] git push failed", "error", err, "output", string(output))
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Pushed branch successfully", "task_id", taskID)
	return nil
}

// GetCurrentBranch returns the current branch name in a worktree.
func (m *GitManager) GetCurrentBranch(taskID string) (string, error) {
	worktreePath := m.worktreePath(taskID)

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch failed: %w", err)
	}

	return string(output), nil
}

// GetDiff returns the git diff for a task's worktree.
// Shows diff between main branch and HEAD (all changes on the feature branch).
func (m *GitManager) GetDiff(taskID string) (string, error) {
	worktreePath := m.worktreePath(taskID)

	// slog.Info("[GIT] GetDiff called", "task_id", taskID, "worktree_path", worktreePath)

	// Get current branch name
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchCmd.Dir = worktreePath
	branchOutput, err := branchCmd.Output()
	if err != nil {
		slog.Error("[GIT] Failed to get branch name", "error", err)
		return "", fmt.Errorf("git branch failed: %w", err)
	}
	currentBranch := strings.TrimSpace(string(branchOutput))

	// Try origin/main first (remote tracking branch)
	cmd := exec.Command("git", "diff", "origin/main", currentBranch)
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// slog.Warn("[GIT] GetDiff origin/main failed, trying main", "error", err)
		// Fallback to local main branch
		cmd = exec.Command("git", "diff", "main", currentBranch)
		cmd.Dir = worktreePath
		output, err = cmd.CombinedOutput()
		if err != nil {
			// slog.Warn("[GIT] GetDiff main failed, trying master", "error", err)
			// Try master branch
			cmd = exec.Command("git", "diff", "master", currentBranch)
			cmd.Dir = worktreePath
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

// PullMainIntoWorktree pulls the latest main into the worktree and merges.
// If there's a merge conflict, returns ErrMergeConflict with the conflicted files.
func (m *GitManager) PullMainIntoWorktree(owner, repo, taskID string) error {
	worktreePath := m.worktreePath(taskID)
	repoPath := m.repoPath(owner, repo)

	slog.Info("[GIT] PullMainIntoWorktree called", "task_id", taskID, "worktree_path", worktreePath)

	// Fetch latest from origin in worktree
	cmd := exec.Command("git", "fetch", "origin", "main")
	cmd.Dir = worktreePath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try master
		cmd = exec.Command("git", "fetch", "origin", "master")
		cmd.Dir = worktreePath
		if output, err := cmd.CombinedOutput(); err != nil {
			slog.Warn("[GIT] Fetch failed", "error", err, "output", string(output))
		}
	}
	slog.Info("[GIT] Fetched latest from origin")

	// Also fetch in main repo to keep it updated
	cmd = exec.Command("git", "fetch", "origin")
	cmd.Dir = repoPath
	_ = cmd.Run()

	// Try to merge origin/main into the worktree
	cmd = exec.Command("git", "merge", "origin/main", "--no-edit")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a merge conflict
		if strings.Contains(string(output), "CONFLICT") || strings.Contains(string(output), "Automatic merge failed") {
			// Get list of conflicted files
			cmd = exec.Command("git", "diff", "--name-only", "--diff-filter=U")
			cmd.Dir = worktreePath
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
				RepoPath:        worktreePath,
			}
		}
		return fmt.Errorf("failed to merge origin/main: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Merged origin/main into worktree successfully")
	return nil
}

// CommitMergeResolution commits after merge conflict resolution.
func (m *GitManager) CommitMergeResolution(taskID, message string) error {
	worktreePath := m.worktreePath(taskID)

	// Stage all changes
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = worktreePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\nOutput: %s", err, string(output))
	}

	// Commit
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = worktreePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, string(output))
	}

	// Push
	cmd = exec.Command("git", "push", "-u", "origin", "HEAD")
	cmd.Dir = worktreePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Committed and pushed merge resolution", "task_id", taskID)
	return nil
}

// AbortMerge aborts an in-progress merge.
func (m *GitManager) AbortMerge(taskID string) error {
	worktreePath := m.worktreePath(taskID)

	cmd := exec.Command("git", "merge", "--abort")
	cmd.Dir = worktreePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git merge --abort failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Aborted merge", "task_id", taskID)
	return nil
}

// MergeToMain merges the task branch to main and pushes.
// Returns the branch name that was merged.
func (m *GitManager) MergeToMain(owner, repo, taskID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	repoPath := m.repoPath(owner, repo)
	worktreePath := m.worktreePath(taskID)

	slog.Info("[GIT] MergeToMain called", "task_id", taskID, "repo_path", repoPath, "worktree_path", worktreePath)

	// Get the branch name from the worktree
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = worktreePath
	branchOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get branch name: %w", err)
	}
	branchName := string(branchOutput)
	// Trim newline
	if len(branchName) > 0 && branchName[len(branchName)-1] == '\n' {
		branchName = branchName[:len(branchName)-1]
	}
	slog.Info("[GIT] Task branch", "branch", branchName)

	// Checkout main in the main repo
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try master
		cmd = exec.Command("git", "checkout", "master")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to checkout main/master: %w\nOutput: %s", err, string(output))
		}
	}
	slog.Info("[GIT] Checked out main branch")

	// Pull latest main
	cmd = exec.Command("git", "pull", "origin", "main")
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try master
		cmd = exec.Command("git", "pull", "origin", "master")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			slog.Warn("[GIT] Pull failed, continuing anyway", "error", err, "output", string(output))
		}
	}
	slog.Info("[GIT] Pulled latest main")

	// Merge the task branch
	cmd = exec.Command("git", "merge", branchName, "--no-edit")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check for merge conflict
		if strings.Contains(string(output), "CONFLICT") || strings.Contains(string(output), "Automatic merge failed") {
			// Abort the merge in main repo
			abortCmd := exec.Command("git", "merge", "--abort")
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
				RepoPath:        worktreePath,
			}
		}
		return "", fmt.Errorf("failed to merge branch %s: %w\nOutput: %s", branchName, err, string(output))
	}
	slog.Info("[GIT] Merged branch", "branch", branchName)

	// Push to origin
	cmd = exec.Command("git", "push", "origin", "main")
	cmd.Dir = repoPath
	if _, err := cmd.CombinedOutput(); err != nil {
		// Try master
		cmd = exec.Command("git", "push", "origin", "master")
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to push to main: %w\nOutput: %s", err, string(output))
		}
	}
	slog.Info("[GIT] Pushed to origin main")

	// Delete the remote branch (optional, don't fail if this errors)
	cmd = exec.Command("git", "push", "origin", "--delete", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("[GIT] Failed to delete remote branch (may not exist)", "branch", branchName, "error", err, "output", string(output))
	} else {
		slog.Info("[GIT] Deleted remote branch", "branch", branchName)
	}

	// Remove the worktree first (branch cannot be deleted while used by worktree)
	if err := os.RemoveAll(worktreePath); err != nil {
		slog.Warn("[GIT] Failed to remove worktree directory", "error", err)
	} else {
		slog.Info("[GIT] Removed worktree directory", "path", worktreePath)
	}

	// Prune worktrees to clean up git's internal state
	cmd = exec.Command("git", "worktree", "prune")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("[GIT] Failed to prune worktrees", "error", err, "output", string(output))
	} else {
		slog.Info("[GIT] Pruned worktrees")
	}

	// Delete the local branch
	cmd = exec.Command("git", "branch", "-d", branchName)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Warn("[GIT] Failed to delete local branch", "branch", branchName, "error", err, "output", string(output))
	} else {
		slog.Info("[GIT] Deleted local branch", "branch", branchName)
	}

	slog.Info("[GIT] MergeToMain completed successfully", "task_id", taskID, "branch", branchName)
	return branchName, nil
}

// RemoveWorktree removes the worktree for a task.
func (m *GitManager) RemoveWorktree(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	worktreePath := m.worktreePath(taskID)
	slog.Info("[GIT] Removing worktree", "task_id", taskID, "path", worktreePath)

	// Check if worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		slog.Info("[GIT] Worktree does not exist, nothing to remove", "task_id", taskID)
		return nil
	}

	// Remove the worktree directory first
	if err := os.RemoveAll(worktreePath); err != nil {
		slog.Error("[GIT] Failed to remove worktree directory", "error", err)
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}

	// Prune worktrees to clean up git's internal state
	// Find the parent repo by looking for any repo in the repos directory
	reposDir := filepath.Join(m.dataDir, "repos")
	_ = filepath.Walk(reposDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".git" {
			parentRepo := filepath.Dir(path)
			cmd := exec.Command("git", "worktree", "prune")
			cmd.Dir = parentRepo
			_ = cmd.Run() // Ignore errors
		}
		return nil
	})

	slog.Info("[GIT] Worktree removed", "task_id", taskID)
	return nil
}
