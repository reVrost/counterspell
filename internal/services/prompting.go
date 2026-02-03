package services

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/revrost/counterspell/internal/prompt"
)

func buildSystemPrompt(repoManager RepoManager, workDir string) string {
	b := prompt.NewBuilder()
	b.AddLine(fmt.Sprintf("You are a coding assistant. Work directory: %s. Be concise. Make changes directly.", workDir))

	if repoManager != nil && repoManager.Kind() == RepoKindJJ {
		agentsPath := filepath.Join(repoManager.RootPath(), "AGENTS.md")
		if err := b.AddFileSection("AGENTS.md", agentsPath); err != nil {
			slog.Warn("[PROMPT] Failed to read AGENTS.md", "error", err, "path", agentsPath)
		}
	}

	return b.String()
}
