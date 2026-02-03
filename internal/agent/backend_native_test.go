package agent

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestNativeBackend_Events(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := NewMockLLMCaller(ctrl)
	mockProvider := &mockLLMProvider{}

	backend, err := NewNativeBackend(
		WithProvider(mockProvider),
	)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	// Inject mock caller into underlying runner
	backend.Runner().llmCaller = mockCaller

	ctx := context.Background()

	t.Run("Event flow for simple text response", func(t *testing.T) {
		mockCaller.EXPECT().
			Stream(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(makeLLMStream([]LLMEvent{
				{Type: LLMContentStart, BlockType: "text", Block: &ContentBlock{Type: "text"}},
				{Type: LLMContentDelta, BlockType: "text", Delta: "Hello, I am the agent."},
				{Type: LLMContentEnd, BlockType: "text"},
				{Type: LLMMessageEnd},
			}), nil)

		stream := backend.Stream(ctx, "Hi")
		events, err := collectStream(stream)
		if err != nil {
			t.Errorf("Stream failed: %v", err)
		}

		if !hasEventType(events, EventContentDelta) {
			t.Errorf("expected content_delta event")
		}
		if !hasEventType(events, EventDone) {
			t.Errorf("expected done event")
		}
	})

	t.Run("Event flow for tool usage", func(t *testing.T) {
		mockCaller.EXPECT().
			Stream(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(makeLLMStream([]LLMEvent{
				{Type: LLMContentStart, BlockType: "tool_use", Block: &ContentBlock{Type: "tool_use", Name: "list_files", ID: "call_123", Input: map[string]any{"path": "."}}},
				{Type: LLMContentEnd, BlockType: "tool_use"},
				{Type: LLMMessageEnd},
			}), nil).Times(1)

		mockCaller.EXPECT().
			Stream(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(makeLLMStream([]LLMEvent{
				{Type: LLMContentStart, BlockType: "text", Block: &ContentBlock{Type: "text"}},
				{Type: LLMContentDelta, BlockType: "text", Delta: "I listed the files."},
				{Type: LLMContentEnd, BlockType: "text"},
				{Type: LLMMessageEnd},
			}), nil).Times(1)

		stream := backend.Stream(ctx, "list files")
		events, err := collectStream(stream)
		if err != nil {
			t.Errorf("Stream failed: %v", err)
		}

		if !hasEventType(events, EventContentEnd) {
			t.Errorf("expected content_end event")
		}
		if !hasBlockType(events, "tool_result") {
			t.Errorf("expected tool_result content")
		}
		if !hasEventType(events, EventDone) {
			t.Errorf("expected done event")
		}
	})
}

func makeLLMStream(events []LLMEvent) *LLMStream {
	eventCh := make(chan LLMEvent, len(events))
	doneCh := make(chan error, 1)
	go func() {
		for _, ev := range events {
			eventCh <- ev
		}
		close(eventCh)
		doneCh <- nil
		close(doneCh)
	}()
	return &LLMStream{Events: eventCh, Done: doneCh}
}

func collectStream(stream *Stream) ([]StreamEvent, error) {
	var events []StreamEvent
	var streamErr error
	for stream.Events != nil || stream.Done != nil {
		select {
		case ev, ok := <-stream.Events:
			if !ok {
				stream.Events = nil
				continue
			}
			events = append(events, ev)
		case err, ok := <-stream.Done:
			if !ok {
				stream.Done = nil
				continue
			}
			streamErr = err
			stream.Done = nil
		}
	}
	return events, streamErr
}

func hasEventType(events []StreamEvent, typ StreamEventType) bool {
	for _, ev := range events {
		if ev.Type == typ {
			return true
		}
	}
	return false
}

func hasBlockType(events []StreamEvent, blockType string) bool {
	for _, ev := range events {
		if ev.Block != nil && ev.Block.Type == blockType {
			return true
		}
		if ev.BlockType == blockType {
			return true
		}
	}
	return false
}
