package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindRepoRootPrefersJJ(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".jj"), 0755))
	sub := filepath.Join(root, "a", "b")
	require.NoError(t, os.MkdirAll(sub, 0755))

	found, kind, err := findRepoRoot(sub)
	require.NoError(t, err)
	require.Equal(t, root, found)
	require.Equal(t, RepoKindJJ, kind)
}

func TestFindRepoRootGit(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0755))
	sub := filepath.Join(root, "nested")
	require.NoError(t, os.MkdirAll(sub, 0755))

	found, kind, err := findRepoRoot(sub)
	require.NoError(t, err)
	require.Equal(t, root, found)
	require.Equal(t, RepoKindGit, kind)
}

func TestFindRepoRootNotFound(t *testing.T) {
	root := t.TempDir()
	_, _, err := findRepoRoot(root)
	require.Error(t, err)
}
