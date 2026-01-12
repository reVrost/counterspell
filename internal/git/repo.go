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

	// Check if repo already exists
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		slog.Info("Repo exists, fetching latest", "owner", owner, "repo", repo)
		if err := m.fetchLatest(repoPath); err != nil {
			return "", fmt.Errorf("failed to fetch: %w", err)
		}
		return repoPath, nil
	}

	// Clone the repo
	slog.Info("Cloning repo", "owner", owner, "repo", repo)
	if err := m.cloneRepo(owner, repo, token, repoPath); err != nil {
		return "", fmt.Errorf("failed to clone: %w", err)
	}

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
	} else {
		cloneURL = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	}

	// Clone with depth 1 for speed, but we need full history for worktrees
	// So we do a regular clone
	cmd := exec.Command("git", "clone", cloneURL, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Cloned repo", "path", destPath)
	return nil
}

// fetchLatest fetches the latest changes from origin.
func (m *RepoManager) fetchLatest(repoPath string) error {
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %w\nOutput: %s", err, string(output))
	}

	// Reset main to origin/main
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = repoPath
	cmd.CombinedOutput() // Ignore error, might be on main already

	cmd = exec.Command("git", "reset", "--hard", "origin/main")
	cmd.Dir = repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		// Try master branch
		cmd = exec.Command("git", "reset", "--hard", "origin/master")
		cmd.Dir = repoPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git reset failed: %w\nOutput: %s", err, string(output))
		}
	}

	slog.Info("Fetched latest", "path", repoPath)
	return nil
}

// CreateWorktree creates an isolated worktree for a task.
// Returns the worktree path.
func (m *RepoManager) CreateWorktree(owner, repo, taskID, branchName string) (string, error) {
	repoPath := m.repoPath(owner, repo)
	worktreePath := m.worktreePath(taskID)

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		slog.Info("Worktree already exists", "path", worktreePath)
		return worktreePath, nil
	}

	// Ensure worktrees directory exists
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktrees dir: %w", err)
	}

	// Create new branch and worktree
	cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreePath)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Branch might already exist, try without -b
		cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
		cmd.Dir = repoPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git worktree add failed: %w\nOutput: %s", err, string(output))
		}
	}

	slog.Info("Created worktree", "task_id", taskID, "path", worktreePath, "branch", branchName)
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

	// Stage all changes
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = worktreePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\nOutput: %s", err, string(output))
	}

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
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, string(output))
	}

	// Push
	cmd = exec.Command("git", "push", "-u", "origin", "HEAD")
	cmd.Dir = worktreePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("Committed and pushed", "task_id", taskID)
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
func (m *RepoManager) GetDiff(taskID string) (string, error) {
	worktreePath := m.worktreePath(taskID)

	cmd := exec.Command("git", "diff", "HEAD")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
