package utils

import (
	"strings"
)

// ExtractTitleFromMarkdown extracts the first heading from markdown content.
// It looks for # heading (H1) first, then ## (H2), etc.
// If no heading is found, it returns the first line or a default.
func ExtractTitleFromMarkdown(content string) string {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Find the heading level and extract text
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	// No heading found, return first non-empty line as title
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			// Limit title to 60 characters
			if len(trimmed) > 60 {
				return trimmed[:57] + "..."
			}
			return trimmed
		}
	}

	return "Untitled Task"
}
