package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitService handles git operations for agents.
type GitService struct {
	dataDir string
}

// NewGitService creates a new git service.
func NewGitService(dataDir string) *GitService {
	return &GitService{dataDir: dataDir}
}

// GitStatus represents git status.
type GitStatus struct {
	Branch        string   `json:"branch"`
	Staged        []string `json:"staged"`
	Modified      []string `json:"modified"`
	Untracked     []string `json:"untracked"`
	AheadCommits  int      `json:"ahead_commits"`
	BehindCommits int      `json:"behind_commits"`
	CurrentCommit string   `json:"current_commit"`
}

// GitBranch represents a git branch.
type GitBranch struct {
	Name   string `json:"name"`
	IsHEAD bool   `json:"is_head"`
}

// GitCommit represents a git commit.
type GitCommit struct {
	Hash      string `json:"hash"`
	Author    string `json:"author"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// GitDiff represents file diff.
type GitDiff struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Changes string `json:"changes"`
}

// Init initializes a git repository.
func (s *GitService) Init(ctx context.Context, path string) error {
	slog.Info("[GIT] Initializing repository", "path", path)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Clone clones a git repository.
func (s *GitService) Clone(ctx context.Context, url, targetPath string) (string, error) {
	slog.Info("[GIT] Cloning repository", "url", url, "target", targetPath)

	fullPath := filepath.Join(s.dataDir, targetPath)

	cmd := exec.CommandContext(ctx, "git", "clone", url, fullPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	return fullPath, nil
}

// Status returns git status.
func (s *GitService) Status(ctx context.Context, path string) (*GitStatus, error) {
	slog.Info("[GIT] Getting status", "path", path)
	fullPath := filepath.Join(s.dataDir, path)

	// Get branch name
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = fullPath
	branchOutput, err := cmd.CombinedOutput()
	branch := strings.TrimSpace(string(branchOutput))
	if err != nil {
		return nil, fmt.Errorf("git branch failed: %w", err)
	}

	// Get current commit
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = fullPath
	commitOutput, err := cmd.CombinedOutput()
	currentCommit := strings.TrimSpace(string(commitOutput))
	if err != nil {
		return nil, fmt.Errorf("git rev-parse failed: %w", err)
	}

	// Get ahead/behind
	cmd = exec.CommandContext(ctx, "git", "rev-list", "--count", "--left-right", "@{u}...HEAD")
	cmd.Dir = fullPath
	abOutput, err := cmd.CombinedOutput()
	var ahead, behind int
	if err == nil {
		parts := strings.Fields(string(abOutput))
		if len(parts) >= 2 {
			_, _ = fmt.Sscanf(parts[0], "%d", &behind)
			_, _ = fmt.Sscanf(parts[1], "%d", &ahead)
		}
	}

	// Get status --porcelain
	cmd = exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = fullPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	var staged, modified, untracked []string
	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		if line == "" {
			continue
		}
		if len(line) < 2 {
			continue
		}

		status := line[:2]
		filename := strings.TrimSpace(line[2:])

		switch {
		case status[0] == 'M' || status[0] == 'A' || status[0] == 'D':
			staged = append(staged, filename)
		case status[1] == 'M' || status[1] == 'D':
			modified = append(modified, filename)
		case status == "??":
			untracked = append(untracked, filename)
		}
	}

	return &GitStatus{
		Branch:        branch,
		Staged:        staged,
		Modified:      modified,
		Untracked:     untracked,
		AheadCommits:  ahead,
		BehindCommits: behind,
		CurrentCommit: currentCommit,
	}, nil
}

// Branches lists all branches.
func (s *GitService) Branches(ctx context.Context, path string) ([]*GitBranch, error) {
	slog.Info("[GIT] Listing branches", "path", path)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "branch", "-a")
	cmd.Dir = fullPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git branch failed: %w", err)
	}

	var branches []*GitBranch
	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		isHEAD := strings.HasPrefix(line, "*")
		name := strings.TrimPrefix(line, "*")
		name = strings.TrimSpace(name)

		branches = append(branches, &GitBranch{
			Name:   name,
			IsHEAD: isHEAD,
		})
	}

	return branches, nil
}

// CreateBranch creates a new branch.
func (s *GitService) CreateBranch(ctx context.Context, path, branchName string) error {
	slog.Info("[GIT] Creating branch", "path", path, "branch", branchName)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "branch", branchName)
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git branch failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CheckoutBranch checks out a branch.
func (s *GitService) CheckoutBranch(ctx context.Context, path, branchName string) error {
	slog.Info("[GIT] Checking out branch", "path", path, "branch", branchName)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "checkout", branchName)
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CheckoutBranchAndCreate creates a new branch and checks it out.
func (s *GitService) CheckoutBranchAndCreate(ctx context.Context, path, branchName string) error {
	slog.Info("[GIT] Creating and checking out branch", "path", path, "branch", branchName)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", branchName)
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout -b failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Add stages files for commit.
func (s *GitService) Add(ctx context.Context, path, files string) error {
	slog.Info("[GIT] Staging files", "path", path, "files", files)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "add", files)
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Commit creates a commit.
func (s *GitService) Commit(ctx context.Context, path, message string) error {
	slog.Info("[GIT] Creating commit", "path", path, "message", message)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Diff gets file changes.
func (s *GitService) Diff(ctx context.Context, path, from, to string) ([]*GitDiff, error) {
	slog.Info("[GIT] Getting diff", "path", path, "from", from, "to", to)
	fullPath := filepath.Join(s.dataDir, path)

	args := []string{"diff"}
	if from != "" {
		args = append(args, from)
	}
	if to != "" {
		args = append(args, to)
	}
	args = append(args, "--name-status")

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = fullPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	var diffs []*GitDiff
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		status := parts[0]
		filename := parts[1]

		diffs = append(diffs, &GitDiff{
			From:    filename,
			To:      filename,
			Changes: status,
		})
	}

	return diffs, nil
}

// Log returns commit history.
func (s *GitService) Log(ctx context.Context, path string, limit int) ([]*GitCommit, error) {
	slog.Info("[GIT] Getting log", "path", path, "limit", limit)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "log", "-n", fmt.Sprintf("%d", limit), "--format=%H|%an|%s|%ci")
	cmd.Dir = fullPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	var commits []*GitCommit
	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 4 {
			continue
		}

		commits = append(commits, &GitCommit{
			Hash:      parts[0],
			Author:    parts[1],
			Message:   parts[2],
			Timestamp: parts[3],
		})
	}

	return commits, nil
}

// Merge merges a branch into the current branch.
func (s *GitService) Merge(ctx context.Context, path, branchName string) error {
	slog.Info("[GIT] Merging branch", "path", path, "branch", branchName)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "merge", branchName)
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git merge failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Pull pulls changes from remote.
func (s *GitService) Pull(ctx context.Context, path string) error {
	slog.Info("[GIT] Pulling changes", "path", path)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Push pushes changes to remote.
func (s *GitService) Push(ctx context.Context, path, branch string) error {
	slog.Info("[GIT] Pushing changes", "path", path, "branch", branch)
	fullPath := filepath.Join(s.dataDir, path)

	args := []string{"push"}
	if branch != "" {
		args = append(args, "origin", branch)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Rebase performs a rebase.
func (s *GitService) Rebase(ctx context.Context, path, onto string) error {
	slog.Info("[GIT] Rebasing", "path", path, "onto", onto)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "rebase", onto)
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git rebase failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// RebaseContinue continues a rebase after resolving conflicts.
func (s *GitService) RebaseContinue(ctx context.Context, path string) error {
	slog.Info("[GIT] Continuing rebase", "path", path)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "rebase", "--continue")
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git rebase --continue failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// RebaseAbort aborts a rebase.
func (s *GitService) RebaseAbort(ctx context.Context, path string) error {
	slog.Info("[GIT] Aborting rebase", "path", path)
	fullPath := filepath.Join(s.dataDir, path)

	cmd := exec.CommandContext(ctx, "git", "rebase", "--abort")
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git rebase --abort failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Reset resets to a commit.
func (s *GitService) Reset(ctx context.Context, path, commit string, hard bool) error {
	slog.Info("[GIT] Resetting", "path", path, "commit", commit, "hard", hard)
	fullPath := filepath.Join(s.dataDir, path)

	args := []string{"reset"}
	if hard {
		args = append(args, "--hard")
	}
	args = append(args, commit)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = fullPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git reset failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// IsGitRepo checks if a directory is a git repository.
func (s *GitService) IsGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}
	return false
}
