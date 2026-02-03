package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// JJManager handles jj workspace operations.
type JJManager struct {
	repoRoot      string
	workspaceBase string
	runner        CommandRunner
	mu            sync.Mutex
}

func NewJJManager(repoRoot string, runner CommandRunner) *JJManager {
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		absRoot = repoRoot
	}
	base := filepath.Dir(absRoot)
	if runner == nil {
		runner = ExecCommandRunner{}
	}
	return &JJManager{repoRoot: absRoot, workspaceBase: base, runner: runner}
}

func (m *JJManager) Kind() RepoKind {
	return RepoKindJJ
}

func (m *JJManager) RootPath() string {
	return m.repoRoot
}

func (m *JJManager) workspaceName(taskID string) string {
	return TaskBranchName(taskID)
}

func (m *JJManager) WorkspacePath(taskID string) string {
	name := m.workspaceName(taskID)
	return filepath.Join(m.workspaceBase, filepath.FromSlash(name))
}

func (m *JJManager) CreateWorkspace(ctx context.Context, taskID, name string) (string, error) {
	workspacePath := m.WorkspacePath(taskID)

	if _, err := os.Stat(workspacePath); err == nil {
		slog.Info("[JJ] Workspace already exists", "task_id", taskID, "path", workspacePath)
		return workspacePath, nil
	}

	if err := os.MkdirAll(filepath.Dir(workspacePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace dir: %w", err)
	}

	args := []string{"workspace", "add", "--name", name, workspacePath}
	if output, err := m.runner.Run(ctx, m.repoRoot, "jj", args...); err != nil {
		return "", fmt.Errorf("jj workspace add failed: %w\nOutput: %s", err, string(output))
	}

	if err := m.setBookmark(ctx, workspacePath, name); err != nil {
		return "", err
	}

	slog.Info("[JJ] Created workspace", "task_id", taskID, "path", workspacePath, "name", name)
	return workspacePath, nil
}

func (m *JJManager) RemoveWorkspace(ctx context.Context, taskID string) error {
	workspacePath := m.WorkspacePath(taskID)
	name := m.workspaceName(taskID)

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		slog.Info("[JJ] Workspace does not exist, nothing to remove", "task_id", taskID)
		return nil
	}

	if output, err := m.runner.Run(ctx, m.repoRoot, "jj", "workspace", "forget", name); err != nil {
		slog.Warn("[JJ] Failed to forget workspace", "task_id", taskID, "error", err, "output", string(output))
	}

	if err := os.RemoveAll(workspacePath); err != nil {
		return fmt.Errorf("failed to remove workspace directory: %w", err)
	}

	slog.Info("[JJ] Removed workspace", "task_id", taskID, "path", workspacePath)
	return nil
}

func (m *JJManager) Commit(ctx context.Context, taskID, message string) error {
	workspacePath := m.WorkspacePath(taskID)
	if output, err := m.runner.Run(ctx, workspacePath, "jj", "describe", "-m", message); err != nil {
		return fmt.Errorf("jj describe failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func (m *JJManager) CommitMergeResolution(ctx context.Context, taskID, message string) error {
	return m.Commit(ctx, taskID, message)
}

func (m *JJManager) AbortMerge(ctx context.Context, taskID string) error {
	return ErrUnsupported{Kind: RepoKindJJ, Op: "abort_merge"}
}

func (m *JJManager) GetCurrentBranch(ctx context.Context, taskID string) (string, error) {
	workspacePath := m.WorkspacePath(taskID)
	bookmark, err := m.currentBookmark(ctx, workspacePath)
	if err != nil {
		return "", err
	}
	if bookmark != "" {
		return bookmark, nil
	}

	changeID, err := m.changeID(ctx, workspacePath)
	if err != nil {
		return "", err
	}
	return changeID, nil
}

func (m *JJManager) PushBranch(ctx context.Context, taskID string) error {
	workspacePath := m.WorkspacePath(taskID)
	bookmark, err := m.currentBookmark(ctx, workspacePath)
	if err != nil {
		return err
	}
	if bookmark == "" {
		bookmark = TaskBranchName(taskID)
		if err := m.setBookmark(ctx, workspacePath, bookmark); err != nil {
			return err
		}
	}

	if output, err := m.runner.Run(ctx, workspacePath, "jj", "git", "push", "--bookmark", bookmark); err != nil {
		return fmt.Errorf("jj git push failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func (m *JJManager) GetDiff(ctx context.Context, taskID string) (string, error) {
	workspacePath := m.WorkspacePath(taskID)
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		slog.Warn("[JJ] Workspace missing, returning empty diff", "task_id", taskID, "path", workspacePath)
		return "", nil
	}

	output, err := m.runner.Run(ctx, workspacePath, "jj", "diff", "-r", "@")
	if err != nil {
		return "", fmt.Errorf("jj diff failed: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

func (m *JJManager) MergeToMain(ctx context.Context, taskID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	workspacePath := m.WorkspacePath(taskID)
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace not found: %s", workspacePath)
	}

	changeID, err := m.changeID(ctx, workspacePath)
	if err != nil {
		return "", err
	}

	if output, err := m.runner.Run(ctx, m.repoRoot, "jj", "squash", "--from", changeID, "--into", "@"); err != nil {
		return "", fmt.Errorf("jj squash failed: %w\nOutput: %s", err, string(output))
	}

	if output, err := m.runner.Run(ctx, m.repoRoot, "jj", "bookmark", "set", "main", "-r", "@"); err != nil {
		return "", fmt.Errorf("jj bookmark set main failed: %w\nOutput: %s", err, string(output))
	}

	if output, err := m.runner.Run(ctx, m.repoRoot, "jj", "git", "push", "--bookmark", "main"); err != nil {
		return "", fmt.Errorf("jj git push failed: %w\nOutput: %s", err, string(output))
	}

	branchName := m.workspaceName(taskID)
	if output, err := m.runner.Run(ctx, m.repoRoot, "jj", "bookmark", "delete", branchName); err != nil {
		slog.Warn("[JJ] Failed to delete task bookmark", "task_id", taskID, "error", err, "output", string(output))
	}

	if err := m.RemoveWorkspace(ctx, taskID); err != nil {
		slog.Warn("[JJ] Failed to remove workspace after merge", "task_id", taskID, "error", err)
	}

	return branchName, nil
}

func (m *JJManager) currentBookmark(ctx context.Context, workspacePath string) (string, error) {
	output, err := m.runner.Run(ctx, workspacePath, "jj", "log", "-r", "@", "-T", `bookmarks.join(",")`)
	if err != nil {
		return "", fmt.Errorf("jj log bookmarks failed: %w\nOutput: %s", err, string(output))
	}
	value := strings.TrimSpace(string(output))
	if value == "" {
		return "", nil
	}
	parts := strings.Split(value, ",")
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			return trimmed, nil
		}
	}
	return "", nil
}

func (m *JJManager) changeID(ctx context.Context, workspacePath string) (string, error) {
	output, err := m.runner.Run(ctx, workspacePath, "jj", "log", "-r", "@", "-T", "change_id")
	if err != nil {
		return "", fmt.Errorf("jj log change_id failed: %w\nOutput: %s", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

func (m *JJManager) setBookmark(ctx context.Context, workspacePath, name string) error {
	if output, err := m.runner.Run(ctx, workspacePath, "jj", "bookmark", "set", name, "-r", "@"); err != nil {
		return fmt.Errorf("jj bookmark set failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
