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

	var receivedEvents []StreamEvent
	callback := func(e StreamEvent) {
		receivedEvents = append(receivedEvents, e)
	}

	backend, err := NewNativeBackend(
		WithProvider(mockProvider),
		WithCallback(callback),
	)
	if err != nil {
		t.Fatalf("Failed to create backend: %v", err)
	}

	// Inject mock caller into underlying runner
	backend.Runner().llmCaller = mockCaller

	ctx := context.Background()

	t.Run("Event flow for simple text response", func(t *testing.T) {
		receivedEvents = nil

		mockCaller.EXPECT().
			Call(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&APIResponse{
				Content: []ContentBlock{{Type: "text", Text: "Hello, I am the agent."}},
			}, nil)

		err := backend.Run(ctx, "Hi")
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}

		expectedTypes := []string{EventUserText, EventText, EventDone}
		if len(receivedEvents) != len(expectedTypes) {
			t.Errorf("expected %d events, got %d", len(expectedTypes), len(receivedEvents))
		}

		for i, eventType := range expectedTypes {
			if i < len(receivedEvents) && receivedEvents[i].Type != eventType {
				t.Errorf("event %d: expected type %s, got %s", i, eventType, receivedEvents[i].Type)
			}
		}
	})

	t.Run("Event flow for tool usage", func(t *testing.T) {
		receivedEvents = nil

		// First call returns a tool use
		mockCaller.EXPECT().
			Call(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&APIResponse{
				Content: []ContentBlock{{
					Type:  "tool_use",
					Name:  "list_files",
					ID:    "call_123",
					Input: map[string]any{"path": "."},
				}},
			}, nil).Times(1)

		// Second call (after tool result) returns a text response
		mockCaller.EXPECT().
			Call(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&APIResponse{
				Content: []ContentBlock{{Type: "text", Text: "I listed the files."}},
			}, nil).Times(1)

		err := backend.Run(ctx, "list files")
		if err != nil {
			t.Errorf("Run failed: %v", err)
		}

		expectedTypes := []string{
			EventUserText,
			EventTool,
			EventToolResult,
			EventText,
			EventDone,
		}

		if len(receivedEvents) != len(expectedTypes) {
			t.Errorf("expected %d events, got %d", len(expectedTypes), len(receivedEvents))
			for i, e := range receivedEvents {
				t.Logf("Event %d: %s", i, e.Type)
			}
		}

		for i, eventType := range expectedTypes {
			if i < len(receivedEvents) && receivedEvents[i].Type != eventType {
				t.Errorf("event %d: expected type %s, got %s", i, eventType, receivedEvents[i].Type)
			}
		}
	})
}
