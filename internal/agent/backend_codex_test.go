package agent

import (
	"bufio"
	"strings"
	"testing"
)

func TestCodexBackend_EventParsing(t *testing.T) {
	var receivedEvents []StreamEvent
	callback := func(e StreamEvent) {
		receivedEvents = append(receivedEvents, e)
	}

	b := &CodexBackend{
		callback: callback,
	}

	rawEvents := []string{
		`{"type":"thread.started","thread_id":"thread_123"}`,
		`{"type":"item.started","item":{"id":"item_1","type":"command_execution","command":"ls","status":"in_progress"}}`,
		`{"type":"item.completed","item":{"id":"item_1","type":"command_execution","command":"ls","status":"completed","stdout":"file1.txt\n"}}`,
		`{"type":"item.completed","item":{"id":"item_2","type":"agent_message","text":"Done."}}`,
		`{"type":"turn.completed"}`,
	}

	input := strings.Join(rawEvents, "\n")
	scanner := bufio.NewScanner(strings.NewReader(input))

	b.parseOutput(scanner)

	expectedTypes := []string{
		"session",
		EventTool,
		EventToolResult,
		EventText,
		EventDone,
	}

	if len(receivedEvents) != len(expectedTypes) {
		t.Fatalf("expected %d events, got %d", len(expectedTypes), len(receivedEvents))
	}

	for i, eventType := range expectedTypes {
		if receivedEvents[i].Type != eventType {
			t.Errorf("event %d: expected type %s, got %s", i, eventType, receivedEvents[i].Type)
		}
	}
}
