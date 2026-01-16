package git

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/revrost/code/counterspell/internal/config"
)

// MultiTenantRepoManager handles shared bare repos and user-isolated worktrees.
// Directory structure:
//   - Shared repos: data/repos/{owner}/{repo}.git (bare clones)
//   - User worktrees: data/workspaces/{user_id}/worktrees/{repo}_{task_id}
type MultiTenantRepoManager struct {
	cfg *config.Config
	mu  sync.Mutex
}

// NewMultiTenantRepoManager creates a new multi-tenant repo manager.
func NewMultiTenantRepoManager(cfg *config.Config) *MultiTenantRepoManager {
	return &MultiTenantRepoManager{cfg: cfg}
}

// bareRepoPath returns the path for a shared bare repo.
func (m *MultiTenantRepoManager) bareRepoPath(owner, repo string) string {
	return filepath.Join(m.cfg.ReposPath(), owner, repo+".git")
}

// worktreePath returns the worktree path for a user's task.
func (m *MultiTenantRepoManager) worktreePath(userID, repo, taskID string) string {
	return filepath.Join(m.cfg.WorkspacesPath(userID), "worktrees", fmt.Sprintf("%s_%s", repo, taskID))
}

// WorktreePath returns the worktree path for a user's task (exported).
func (m *MultiTenantRepoManager) WorktreePath(userID, repo, taskID string) string {
	return m.worktreePath(userID, repo, taskID)
}

// RepoPath returns the bare repo path (exported).
func (m *MultiTenantRepoManager) RepoPath(owner, repo string) string {
	return m.bareRepoPath(owner, repo)
}

// EnsureRepo ensures the shared bare repo exists, cloning if necessary.
// Returns the bare repo path.
func (m *MultiTenantRepoManager) EnsureRepo(owner, repo, token string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bareRepoPath := m.bareRepoPath(owner, repo)
	slog.Info("[GIT] EnsureRepo called", "owner", owner, "repo", repo, "bare_repo_path", bareRepoPath)

	// Check if bare repo already exists
	if _, err := os.Stat(filepath.Join(bareRepoPath, "HEAD")); err == nil {
		slog.Info("[GIT] Bare repo exists, fetching latest", "owner", owner, "repo", repo)
		if err := m.fetchBareRepo(bareRepoPath, token, owner, repo); err != nil {
			slog.Error("[GIT] Failed to fetch", "error", err)
			return "", fmt.Errorf("failed to fetch: %w", err)
		}
		return bareRepoPath, nil
	}

	// Clone as bare repo
	slog.Info("[GIT] Cloning bare repo", "owner", owner, "repo", repo, "dest", bareRepoPath)
	if err := m.cloneBareRepo(owner, repo, token, bareRepoPath); err != nil {
		slog.Error("[GIT] Failed to clone", "error", err)
		return "", fmt.Errorf("failed to clone: %w", err)
	}

	slog.Info("[GIT] Bare repo ready", "path", bareRepoPath)
	return bareRepoPath, nil
}

// cloneBareRepo clones a repo as a bare repository.
func (m *MultiTenantRepoManager) cloneBareRepo(owner, repo, token, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var cloneURL string
	if token != "" {
		cloneURL = fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)
	} else {
		cloneURL = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	}

	cmd := exec.Command("git", "clone", "--bare", cloneURL, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("[GIT] Bare clone failed", "error", err, "output", string(output))
		return fmt.Errorf("git clone --bare failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Bare repo cloned successfully", "path", destPath)
	return nil
}

// fetchBareRepo fetches the latest changes into the bare repo.
func (m *MultiTenantRepoManager) fetchBareRepo(bareRepoPath, token, owner, repo string) error {
	// Update remote URL with token if provided
	if token != "" {
		newURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)
		cmd := exec.Command("git", "remote", "set-url", "origin", newURL)
		cmd.Dir = bareRepoPath
		if output, err := cmd.CombinedOutput(); err != nil {
			slog.Error("[GIT] Failed to update remote URL", "error", err, "output", string(output))
		}
	}

	cmd := exec.Command("git", "fetch", "--all", "--prune")
	cmd.Dir = bareRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("[GIT] Fetch failed", "error", err, "output", string(output))
		return fmt.Errorf("git fetch failed: %w\nOutput: %s", err, string(output))
	}

	slog.Info("[GIT] Bare repo fetched successfully", "path", bareRepoPath)
	return nil
}

// CreateWorktree creates an isolated worktree for a user's task.
// Returns the worktree path.
func (m *MultiTenantRepoManager) CreateWorktree(userID, owner, repo, taskID, branchName string) (string, error) {
	bareRepoPath := m.bareRepoPath(owner, repo)
	worktreePath := m.worktreePath(userID, repo, taskID)

	slog.Info("[GIT] Creating worktree",
		"user_id", userID,
		"task_id", taskID,
		"bare_repo", bareRepoPath,
		"worktree", worktreePath,
		"branch", branchName)

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		slog.Info("[GIT] Worktree already exists", "path", worktreePath)
		return worktreePath, nil
	}

	// Ensure worktrees directory exists
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktrees dir: %w", err)
	}

	// Determine the base reference (usually main or master)
	baseRef := m.getDefaultBranch(bareRepoPath)

	// Create worktree from bare repo with new branch
	cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreePath, baseRef)
	cmd.Dir = bareRepoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Warn("[GIT] First attempt failed, trying without -b", "error", err, "output", string(output))
		// Branch might already exist, try without -b
		cmd = exec.Command("git", "worktree", "add", worktreePath, branchName)
		cmd.Dir = bareRepoPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			slog.Error("[GIT] Worktree creation failed", "error", err, "output", string(output))
			return "", fmt.Errorf("git worktree add failed: %w\nOutput: %s", err, string(output))
		}
	}

	slog.Info("[GIT] Worktree created successfully", "path", worktreePath)
	return worktreePath, nil
}

// getDefaultBranch determines the default branch of a repo.
func (m *MultiTenantRepoManager) getDefaultBranch(bareRepoPath string) string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = bareRepoPath
	output, err := cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(output))
		// refs/remotes/origin/main -> main
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// Fallback: check if main or master exists
	for _, branch := range []string{"main", "master"} {
		cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
		cmd.Dir = bareRepoPath
		if cmd.Run() == nil {
			return branch
		}
	}

	return "main" // Default fallback
}

// RemoveWorktree removes a worktree.
func (m *MultiTenantRepoManager) RemoveWorktree(userID, owner, repo, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bareRepoPath := m.bareRepoPath(owner, repo)
	worktreePath := m.worktreePath(userID, repo, taskID)

	slog.Info("[GIT] Removing worktree", "path", worktreePath)

	// Check if worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		slog.Info("[GIT] Worktree does not exist, nothing to remove")
		return nil
	}

	// Remove the worktree directory
	if err := os.RemoveAll(worktreePath); err != nil {
		slog.Error("[GIT] Failed to remove worktree directory", "error", err)
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}

	// Prune worktrees in the bare repo
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = bareRepoPath
	_ = cmd.Run()

	slog.Info("[GIT] Worktree removed", "path", worktreePath)
	return nil
}

// GetWorktreeBranch returns the current branch name in a worktree.
func (m *MultiTenantRepoManager) GetWorktreeBranch(worktreePath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = worktreePath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git branch failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CommitAndPush commits all changes in a worktree and pushes.
func (m *MultiTenantRepoManager) CommitAndPush(worktreePath, message string) error {
	// Stage all changes
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = worktreePath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\nOutput: %s", err, string(output))
	}

	// Check if there are changes to commit
	cmd = exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = worktreePath
	if cmd.Run() == nil {
		slog.Info("[GIT] No changes to commit", "path", worktreePath)
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

	slog.Info("[GIT] Committed and pushed successfully", "path", worktreePath)
	return nil
}

// GetDiff returns the diff between the worktree and the default branch.
func (m *MultiTenantRepoManager) GetDiff(worktreePath string) (string, error) {
	// Try origin/main first
	cmd := exec.Command("git", "diff", "origin/main...HEAD")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try origin/master
		cmd = exec.Command("git", "diff", "origin/master...HEAD")
		cmd.Dir = worktreePath
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git diff failed: %w\nOutput: %s", err, string(output))
		}
	}
	return string(output), nil
}

// PullMain pulls the latest main/master into the worktree.
func (m *MultiTenantRepoManager) PullMain(worktreePath string) error {
	// Fetch
	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = worktreePath
	_ = cmd.Run()

	// Try merging origin/main
	cmd = exec.Command("git", "merge", "origin/main", "--no-edit")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check for conflicts
		if strings.Contains(string(output), "CONFLICT") {
			return m.parseMergeConflict(worktreePath, string(output))
		}
		// Try origin/master
		cmd = exec.Command("git", "merge", "origin/master", "--no-edit")
		cmd.Dir = worktreePath
		output, err = cmd.CombinedOutput()
		if err != nil {
			if strings.Contains(string(output), "CONFLICT") {
				return m.parseMergeConflict(worktreePath, string(output))
			}
			return fmt.Errorf("git merge failed: %w\nOutput: %s", err, string(output))
		}
	}

	return nil
}

func (m *MultiTenantRepoManager) parseMergeConflict(worktreePath, output string) error {
	// Get list of conflicted files
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	cmd.Dir = worktreePath
	conflictOutput, _ := cmd.Output()

	var files []string
	for _, f := range strings.Split(strings.TrimSpace(string(conflictOutput)), "\n") {
		if f != "" {
			files = append(files, f)
		}
	}

	return &ErrMergeConflict{
		ConflictedFiles: files,
		RepoPath:        worktreePath,
	}
}

// AbortMerge aborts an in-progress merge.
func (m *MultiTenantRepoManager) AbortMerge(worktreePath string) error {
	cmd := exec.Command("git", "merge", "--abort")
	cmd.Dir = worktreePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git merge --abort failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
