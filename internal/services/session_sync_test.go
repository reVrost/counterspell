package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCodexSessionJSONLNoMessages(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.jsonl")
	content := []byte("{\"session_id\":\"sess-123\"}\n{\"foo\":\"bar\"}\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	sessionID, messages, err := parseCodexSession(path)
	if err != nil {
		t.Fatalf("parseCodexSession: %v", err)
	}
	if sessionID != "sess-123" {
		t.Fatalf("expected session id sess-123, got %q", sessionID)
	}
	if len(messages) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(messages))
	}
}
