package git

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// RepoManager handles cloning and updating repositories.
type RepoManager struct {
	dataDir string
	mu      sync.Mutex
}

// NewRepoManager creates a new repo manager.
// dataDir is the base directory for storing repos (e.g., "./data")
func NewRepoManager(dataDir string) *RepoManager {
	// Convert to absolute path to ensure worktrees are created at correct level
	absDir, err := filepath.Abs(dataDir)
	if err != nil {
		absDir = dataDir // fallback if conversion fails
	}
	return &RepoManager{dataDir: absDir}
}

// repoPath returns the local path for a given owner/repo.
func (m *RepoManager) repoPath(owner, repo string) string {
	return filepath.Join(m.dataDir, "repos", owner, repo)
}

// worktreePath returns the worktree path for a given task.
func (m *RepoManager) worktreePath(taskID string) string {
	return filepath.Join(m.dataDir, "worktrees", "task-"+taskID)
}

// EnsureRepo clones or updates a repository.
// Returns the local repo path.
func (m *RepoManager) EnsureRepo(owner, repo, token string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	repoPath := m.repoPath(owner, repo)
	slog.Info("[GIT] EnsureRepo called", "owner", owner, "repo", repo, "repo_path", repoPath)

	// Check if repo already exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		slog.Info("[GIT] Repo exists, fetching latest", "owner", owner, "repo", repo, "token_provided", token != "")
		if err := m.fetchLatest(repoPath, token, owner, repo); err != nil {
			slog.Error("[GIT] Failed to fetch", "error", err)
			return "", fmt.Errorf("failed to fetch: %w", err)
		}
		slog.Info("[GIT] Repo fetched successfully", "path", repoPath)
		return repoPath, nil
	}

	// Clone the repo
	slog.Info("[GIT] Cloning repo", "owner", owner, "repo", repo, "dest", repoPath)
	if err := m.cloneRepo(owner, repo, token, repoPath); err != nil {
		slog.Error("[GIT] Failed to clone", "error", err)
		return "", fmt.Errorf("failed to clone: %w", err)
	}

	slog.Info("[GIT] Repo cloned successfully", "path", repoPath)
	return repoPath, nil
}

// cloneRepo performs the actual clone.
func (m *RepoManager) cloneRepo(owner, repo, token, destPath string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Build clone URL with token for auth
	var cloneURL string
	if token != "" {
		cloneURL = fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)
		slog.Info("[GIT] Using authenticated clone URL", "token_length", len(token))
	} else {
		cloneURL = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
		slog.Warn("[GIT] No token provided, cloning without auth")
	}

	slog.Info("[GIT] Cloning repo", "owner", owner, "repo", repo, "dest", destPath)

	// Clone with depth 1 for speed, but we need full history for worktrees
	// So we do a regular clone
	cmd := exec.Command("git", "clone", cloneURL, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("[GIT] Clone command failed", "error", err, "output", string(output))
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Cloned repo successfully", "path", destPath)
	return nil
}

// fetchLatest fetches the latest changes from origin.
func (m *RepoManager) fetchLatest(repoPath, token, owner, repo string) error {
	slog.Info("[GIT] Fetching latest changes", "path", repoPath, "has_token", token != "")

	// Update remote URL with current token if provided
	if token != "" {
		newURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)
		slog.Info("[GIT] Updating remote URL with token", "token_length", len(token))
		cmd := exec.Command("git", "remote", "set-url", "origin", newURL)
		cmd.Dir = repoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			slog.Error("[GIT] Failed to update remote URL", "error", err, "output", string(output))
			return fmt.Errorf("failed to update remote URL: %w\nOutput: %s", err, string(output))
		}
		slog.Info("[GIT] Remote URL updated successfully")
	}

	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("[GIT] git fetch failed", "error", err, "output", string(output))
		return fmt.Errorf("git fetch failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Fetch successful, resetting to origin/main")

	// Reset main to origin/main
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	cmd.CombinedOutput() // Ignore error, might be on main already

	cmd = exec.Command("git", "reset", "--hard", "origin/main")
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		// Try master branch
		slog.Info("[GIT] Trying master branch instead", "path", repoPath)
		cmd = exec.Command("git", "reset", "--hard", "origin/master")
		cmd.Dir = repoPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			slog.Error("[GIT] git reset failed", "error", err, "output", string(output))
			return fmt.Errorf("git reset failed: %w\nOutput: %s", err, string(output))
		}
	}

	slog.Info("[GIT] Fetched and reset successfully", "path", repoPath)
	return nil
}

// CreateWorktree creates an isolated worktree for a task.
// Returns the worktree path.
func (m *RepoManager) CreateWorktree(owner, repo, taskID, branchName string) (string, error) {
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
func (m *RepoManager) CleanupWorktree(owner, repo, taskID string) error {
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

// CommitAndPush commits changes and pushes the branch.
func (m *RepoManager) CommitAndPush(taskID, message string) error {
	worktreePath := m.worktreePath(taskID)

	slog.Info("[GIT] CommitAndPush called", "task_id", taskID, "worktree_path", worktreePath)

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
	slog.Info("[GIT] git commit successful")

	// Push
	cmd = exec.Command("git", "push", "-u", "origin", "HEAD")
	cmd.Dir = worktreePath
	slog.Info("[GIT] Executing: git push -u origin HEAD", "dir", worktreePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		slog.Error("[GIT] git push failed", "error", err, "output", string(output))
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Committed and pushed successfully", "task_id", taskID)
	return nil
}

// GetCurrentBranch returns the current branch name in a worktree.
func (m *RepoManager) GetCurrentBranch(taskID string) (string, error) {
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
// After commit, shows diff between HEAD~1 and HEAD (the committed changes).
func (m *RepoManager) GetDiff(taskID string) (string, error) {
	worktreePath := m.worktreePath(taskID)

	slog.Info("[GIT] GetDiff called", "task_id", taskID, "worktree_path", worktreePath)

	// First try uncommitted changes
	cmd := exec.Command("git", "diff", "HEAD")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("[GIT] GetDiff failed", "error", err, "output", string(output))
		return "", fmt.Errorf("git diff failed: %w\nOutput: %s", err, string(output))
	}

	// If no uncommitted changes, get diff of last commit
	if len(output) == 0 {
		slog.Info("[GIT] No uncommitted changes, getting last commit diff", "task_id", taskID)
		cmd = exec.Command("git", "diff", "HEAD~1", "HEAD")
		cmd.Dir = worktreePath
		output, err = cmd.CombinedOutput()
		if err != nil {
			slog.Warn("[GIT] GetDiff HEAD~1 failed (may be first commit)", "error", err)
			// Try show for first commit case
			cmd = exec.Command("git", "show", "--format=", "HEAD")
			cmd.Dir = worktreePath
			output, _ = cmd.CombinedOutput()
		}
	}

	slog.Info("[GIT] GetDiff successful", "task_id", taskID, "diff_size", len(output))
	return string(output), nil
}

// RemoveWorktree removes the worktree for a task.
func (m *RepoManager) RemoveWorktree(taskID string) error {
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
	filepath.Walk(reposDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".git" {
			parentRepo := filepath.Dir(path)
			cmd := exec.Command("git", "worktree", "prune")
			cmd.Dir = parentRepo
			cmd.Run() // Ignore errors
		}
		return nil
	})

	slog.Info("[GIT] Worktree removed", "task_id", taskID)
	return nil
}
