package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// RepoKind identifies the VCS backend in use.
type RepoKind string

const (
	RepoKindGit RepoKind = "git"
	RepoKindJJ  RepoKind = "jj"
)

// RepoManager abstracts git/jj workspace operations.
type RepoManager interface {
	Kind() RepoKind
	RootPath() string
	WorkspacePath(taskID string) string
	CreateWorkspace(ctx context.Context, taskID, name string) (string, error)
	RemoveWorkspace(ctx context.Context, taskID string) error
	Commit(ctx context.Context, taskID, message string) error
	CommitMergeResolution(ctx context.Context, taskID, message string) error
	AbortMerge(ctx context.Context, taskID string) error
	GetCurrentBranch(ctx context.Context, taskID string) (string, error)
	PushBranch(ctx context.Context, taskID string) error
	GetDiff(ctx context.Context, taskID string) (string, error)
	MergeToMain(ctx context.Context, taskID string) (string, error)
}

// TaskBranchName returns the branch/workspace name for a task.
func TaskBranchName(taskID string) string {
	return "agent/task-" + taskID
}

// ErrRepoRootNotFound is returned when no repo root is discovered.
type ErrRepoRootNotFound struct {
	Start string
}

func (e ErrRepoRootNotFound) Error() string {
	return fmt.Sprintf("no repo root found from %s", e.Start)
}

// ErrUnsupported indicates an unsupported operation for a repo kind.
type ErrUnsupported struct {
	Kind RepoKind
	Op   string
}

func (e ErrUnsupported) Error() string {
	return fmt.Sprintf("%s: %s is unsupported", e.Kind, e.Op)
}

// NewRepoManager detects the repo kind from the current working directory.
func NewRepoManager(dataDir string) (RepoManager, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	root, kind, err := findRepoRoot(cwd)
	if err != nil {
		return nil, err
	}
	switch kind {
	case RepoKindJJ:
		return NewJJManager(root, ExecCommandRunner{}), nil
	case RepoKindGit:
		return NewGitManager(root, dataDir), nil
	default:
		return nil, fmt.Errorf("unsupported repo kind: %s", kind)
	}
}

func findRepoRoot(start string) (string, RepoKind, error) {
	dir := start
	for {
		jjPath := filepath.Join(dir, ".jj")
		if pathExists(jjPath) {
			return dir, RepoKindJJ, nil
		}
		gitPath := filepath.Join(dir, ".git")
		if pathExists(gitPath) {
			return dir, RepoKindGit, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", ErrRepoRootNotFound{Start: start}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
