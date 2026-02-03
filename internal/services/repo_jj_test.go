package services

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestJJGetCurrentBranchPrefersBookmark(t *testing.T) {
	ctrl := gomock.NewController(t)
	runner := NewMockCommandRunner(ctrl)
	jm := NewJJManager(t.TempDir(), runner)
	taskID := "task-1"
	workspacePath := jm.WorkspacePath(taskID)

	runner.EXPECT().Run(gomock.Any(), workspacePath, "jj", "log", "-r", "@", "-T", `bookmarks.join(",")`).Return([]byte("feat-one\n"), nil)

	branch, err := jm.GetCurrentBranch(context.Background(), taskID)
	require.NoError(t, err)
	require.Equal(t, "feat-one", branch)
}

func TestJJGetCurrentBranchFallsBackToChangeID(t *testing.T) {
	ctrl := gomock.NewController(t)
	runner := NewMockCommandRunner(ctrl)
	jm := NewJJManager(t.TempDir(), runner)
	taskID := "task-2"
	workspacePath := jm.WorkspacePath(taskID)

	gomock.InOrder(
		runner.EXPECT().Run(gomock.Any(), workspacePath, "jj", "log", "-r", "@", "-T", `bookmarks.join(",")`).Return([]byte("\n"), nil),
		runner.EXPECT().Run(gomock.Any(), workspacePath, "jj", "log", "-r", "@", "-T", "change_id").Return([]byte("abc123\n"), nil),
	)

	branch, err := jm.GetCurrentBranch(context.Background(), taskID)
	require.NoError(t, err)
	require.Equal(t, "abc123", branch)
}

func TestJJMergeToMain(t *testing.T) {
	ctrl := gomock.NewController(t)
	runner := NewMockCommandRunner(ctrl)
	repoRoot := t.TempDir()
	jm := NewJJManager(repoRoot, runner)
	taskID := "task-3"
	workspacePath := jm.WorkspacePath(taskID)
	require.NoError(t, os.MkdirAll(workspacePath, 0755))
	changeID := "zzzz1234"
	branchName := TaskBranchName(taskID)

	gomock.InOrder(
		runner.EXPECT().Run(gomock.Any(), workspacePath, "jj", "log", "-r", "@", "-T", "change_id").Return([]byte(changeID+"\n"), nil),
		runner.EXPECT().Run(gomock.Any(), repoRoot, "jj", "squash", "--from", changeID, "--into", "@").Return([]byte(""), nil),
		runner.EXPECT().Run(gomock.Any(), repoRoot, "jj", "bookmark", "set", "main", "-r", "@").Return([]byte(""), nil),
		runner.EXPECT().Run(gomock.Any(), repoRoot, "jj", "git", "push", "--bookmark", "main").Return([]byte(""), nil),
		runner.EXPECT().Run(gomock.Any(), repoRoot, "jj", "bookmark", "delete", branchName).Return([]byte(""), nil),
		runner.EXPECT().Run(gomock.Any(), repoRoot, "jj", "workspace", "forget", branchName).Return([]byte(""), nil),
	)

	merged, err := jm.MergeToMain(context.Background(), taskID)
	require.NoError(t, err)
	require.Equal(t, branchName, merged)
}
