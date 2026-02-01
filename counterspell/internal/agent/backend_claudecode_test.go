package agent

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
)

func TestClaudeCodeBackend_EnvVariables(t *testing.T) {
	ctx := context.Background()

	dummyPath := "/tmp/dummy-claude"
	os.WriteFile(dummyPath, []byte("#!/bin/sh\nexit 0"), 0755)
	defer os.Remove(dummyPath)

	tests := []struct {
		name     string
		model    string
		wantEnvs []string
	}{
		{
			name:  "GLM-4.7 model sets extra envs",
			model: "glm-4.7",
			wantEnvs: []string{
				"ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-4.7",
				"ANTHROPIC_DEFAULT_SONNET_MODEL=glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL=glm-4.7",
			},
		},
		{
			name:  "Other model does not set extra envs",
			model: "claude-3-sonnet",
			wantEnvs: []string{
				"ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-4.7",
				"ANTHROPIC_DEFAULT_SONNET_MODEL=glm-4.7",
				"ANTHROPIC_DEFAULT_OPUS_MODEL=glm-4.7",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewClaudeCodeBackend(
				WithBinaryPath(dummyPath),
				WithModel(tt.model),
			)
			if err != nil {
				t.Fatalf("Failed to create backend: %v", err)
			}

			cmd, err := b.buildCmd(ctx, "test prompt")
			if err != nil {
				t.Fatalf("Failed to build command: %v", err)
			}

			if tt.model == "glm-4.7" {
				for _, want := range tt.wantEnvs {
					found := false
					for _, env := range cmd.Env {
						if env == want {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected env %s not found", want)
					}
				}
			} else {
				for _, notWant := range tt.wantEnvs {
					for _, env := range cmd.Env {
						if env == notWant {
							t.Errorf("unexpected env %s found", notWant)
						}
					}
				}
			}
		})
	}
}

func TestClaudeCodeBackend_EventParsing(t *testing.T) {
	var receivedEvents []StreamEvent
	callback := func(e StreamEvent) {
		receivedEvents = append(receivedEvents, e)
	}

	b := &ClaudeCodeBackend{
		callback: callback,
	}

	rawEvents := []string{
		`{"type": "user", "message": {"content": [{"type": "text", "text": "hello"}]}}`,
		`{"type": "assistant", "message": {"content": [{"type": "text", "text": "Hi there!"}, {"type": "tool_use", "name": "ls", "id": "1", "input": {"path": "."}}]}}`,
		`{"type": "tool_result", "content": "file1.txt", "tool_use_id": "1"}`,
		`{"type": "result", "result": "Task completed", "is_error": false}`,
	}

	input := strings.Join(rawEvents, "\n")
	scanner := bufio.NewScanner(strings.NewReader(input))

	b.parseOutput(scanner)

	expectedTypes := []string{
		EventText,       // Assistant text
		EventTool,       // Tool use
		EventToolResult, // Tool result
		EventDone,       // Final result
	}

	if len(receivedEvents) != len(expectedTypes) {
		t.Errorf("expected %d events, got %d", len(expectedTypes), len(receivedEvents))
	}

	for i, eventType := range expectedTypes {
		if i < len(receivedEvents) && receivedEvents[i].Type != eventType {
			t.Errorf("event %d: expected type %s, got %s", i, eventType, receivedEvents[i].Type)
		}
	}
}
