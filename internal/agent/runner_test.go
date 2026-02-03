package agent

import (
	"context"
	"testing"

	"github.com/revrost/counterspell/internal/agent/tools"
	"github.com/revrost/counterspell/internal/llm"
	"go.uber.org/mock/gomock"
)

func TestRunner_EmptyUserMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := NewMockLLMCaller(ctrl)
	mockProvider := &mockLLMProvider{}

	r := NewRunner(mockProvider, ".")
	r.llmCaller = mockCaller

	ctx := context.Background()

	t.Run("New run with empty message should error", func(t *testing.T) {
		err := r.runWithMessage(ctx, "", false, make(chan StreamEvent, 1), make(chan []tools.TodoItem, 1))
		if err == nil {
			t.Error("expected error for empty message on new run, got nil")
		}
	})

	t.Run("Continuation with empty message should NOT add a user message but should call LLM", func(t *testing.T) {
		r.messageHistory = []Message{
			{Role: "user", Content: []ContentBlock{{Type: "text", Text: "initial"}}},
			{Role: "assistant", Content: []ContentBlock{{Type: "text", Text: "response"}}},
		}

		// Expect LLM call with NO new user message added to the input
		mockCaller.EXPECT().
			Stream(gomock.Any(), gomock.Len(2), gomock.Any(), gomock.Any()).
			Return(makeLLMStream([]LLMEvent{
				{Type: LLMContentStart, BlockType: "text", Block: &ContentBlock{Type: "text"}},
				{Type: LLMContentDelta, BlockType: "text", Delta: "next"},
				{Type: LLMContentEnd, BlockType: "text"},
				{Type: LLMMessageEnd},
			}), nil)

		err := r.runWithMessage(ctx, "", true, make(chan StreamEvent, 64), make(chan []tools.TodoItem, 1))
		if err != nil {
			t.Errorf("expected no error for empty message on continuation, got %v", err)
		}

		// Resulting history should have: initial user, initial assistant, NEW assistant response
		// total 3 messages, no empty user message in between.
		if len(r.messageHistory) != 3 {
			t.Errorf("expected 3 messages, got %d", len(r.messageHistory))
		}

		for i, msg := range r.messageHistory {
			if msg.Role == "user" {
				for _, block := range msg.Content {
					if block.Type == "text" && block.Text == "" {
						t.Errorf("message %d has empty user text", i)
					}
				}
			}
		}
	})
}

func TestRunner_MessagePersistence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := NewMockLLMCaller(ctrl)
	mockProvider := &mockLLMProvider{}

	r := NewRunner(mockProvider, ".")
	r.llmCaller = mockCaller

	ctx := context.Background()

	mockCaller.EXPECT().
		Stream(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(makeLLMStream([]LLMEvent{
			{Type: LLMContentStart, BlockType: "text", Block: &ContentBlock{Type: "text"}},
			{Type: LLMContentDelta, BlockType: "text", Delta: "response"},
			{Type: LLMContentEnd, BlockType: "text"},
			{Type: LLMMessageEnd},
		}), nil)

	err := r.Run(ctx, "hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Should have 2 messages in history: user "hello" and assistant "response"
	if len(r.messageHistory) != 2 {
		t.Errorf("expected 2 messages in history, got %d", len(r.messageHistory))
	}
	if r.messageHistory[0].Role != "user" || r.messageHistory[0].Content[0].Text != "hello" {
		t.Errorf("first message should be user 'hello', got %v", r.messageHistory[0])
	}
	if r.messageHistory[1].Role != "assistant" || r.messageHistory[1].Content[0].Text != "response" {
		t.Errorf("second message should be assistant 'response', got %v", r.messageHistory[1])
	}
}

// Minimal mock provider just to satisfy NewRunner
type mockLLMProvider struct {
	llm.Provider
}

func (m *mockLLMProvider) Type() string          { return "anthropic" }
func (m *mockLLMProvider) SetModel(model string) {}
func (m *mockLLMProvider) Model() string         { return "test-model" }
func (m *mockLLMProvider) APIURL() string        { return "" }
func (m *mockLLMProvider) APIKey() string        { return "" }
func (m *mockLLMProvider) APIVersion() string    { return "" }
