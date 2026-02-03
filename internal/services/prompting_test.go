package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildSystemPromptIncludesAgentsForJJ(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".jj"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("TEST-AGENTS"), 0644))

	manager := NewJJManager(root, nil)
	prompt := buildSystemPrompt(manager, "/tmp/workdir")
	require.Contains(t, prompt, "TEST-AGENTS")
	require.Contains(t, prompt, "AGENTS.md")
}
