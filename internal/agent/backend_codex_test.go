package agent

import (
	"bufio"
	"context"
	"strings"
	"testing"
)

func TestCodexBackend_EventParsing(t *testing.T) {
	b := &CodexBackend{}
	events := make(chan StreamEvent, 64)
	b.setStream(context.Background(), events)
	defer b.clearStream()

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

	var receivedEvents []StreamEvent
	for len(events) > 0 {
		receivedEvents = append(receivedEvents, <-events)
	}

	if !hasEventType(receivedEvents, EventSession) {
		t.Fatalf("expected session event")
	}
	if !hasBlockType(receivedEvents, "tool_use") {
		t.Errorf("expected tool_use content block")
	}
	if !hasBlockType(receivedEvents, "tool_result") {
		t.Errorf("expected tool_result content block")
	}
	if !hasBlockType(receivedEvents, "text") {
		t.Errorf("expected text content block")
	}
	if !hasEventType(receivedEvents, EventDone) {
		t.Fatalf("expected done event")
	}
}
