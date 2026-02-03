package services

import (
	"strings"

	"github.com/revrost/counterspell/internal/models"
)

var codexSetupMarkers = []string{
	"agents.md",
	"<environment_context>",
	"<collaboration_mode>",
	"<instructions>",
	"<permissions instructions>",
}

func filterCodexSetupMessages(messages []models.SessionMessage) []models.SessionMessage {
	if len(messages) == 0 {
		return messages
	}
	filtered := make([]models.SessionMessage, 0, len(messages))
	for _, msg := range messages {
		if isCodexSetupMessage(msg.Role, msg.Kind, valueOrEmpty(msg.Content)) {
			continue
		}
		filtered = append(filtered, msg)
	}
	return filtered
}

func filterCodexSetupImportedMessages(messages []importedMessage) []importedMessage {
	if len(messages) == 0 {
		return messages
	}
	filtered := make([]importedMessage, 0, len(messages))
	for _, msg := range messages {
		if isCodexSetupMessage(msg.Role, msg.Kind, msg.Content) {
			continue
		}
		filtered = append(filtered, msg)
	}
	return filtered
}

func isCodexSetupMessage(role, kind, content string) bool {
	if strings.ToLower(kind) == "setup" {
		return true
	}
	if strings.ToLower(role) == "user" && isCodexSetupContent(content) {
		return true
	}
	return false
}

func isCodexSetupContent(content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}
	lower := strings.ToLower(content)
	for _, marker := range codexSetupMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
