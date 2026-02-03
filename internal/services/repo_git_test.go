package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitManagerGetDiffMissingWorkspace(t *testing.T) {
	gm := NewGitManager(t.TempDir(), t.TempDir())
	diff, err := gm.GetDiff(context.Background(), "missing-task")
	require.NoError(t, err)
	require.Empty(t, diff)
}
