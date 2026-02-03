package services

import (
	"testing"

	"github.com/revrost/counterspell/internal/agent"
	"github.com/stretchr/testify/require"
)

func TestStreamAssembler_BuildsBlocks(t *testing.T) {
	assembler := newStreamAssembler()

	msgID := "msg-1"
	events := []agent.StreamEvent{
		{Type: agent.EventMessageStart, MessageID: msgID, Role: "assistant"},
		{Type: agent.EventContentStart, MessageID: msgID, BlockType: "tool_use", Block: &agent.ContentBlock{Type: "tool_use", Name: "ls", ID: "tool-1"}},
		{Type: agent.EventContentDelta, MessageID: msgID, BlockType: "tool_use", Delta: `{"path":"."}`},
		{Type: agent.EventContentEnd, MessageID: msgID, BlockType: "tool_use"},
		{Type: agent.EventMessageEnd, MessageID: msgID, Role: "assistant"},
	}

	var msg *streamMessage
	var ok bool
	for _, ev := range events {
		msg, ok = assembler.Apply(ev)
	}

	require.True(t, ok)
	require.NotNil(t, msg)
	require.Len(t, msg.blocks, 1)
	require.Equal(t, "tool_use", msg.blocks[0].Type)
	require.Equal(t, "ls", msg.blocks[0].Name)
	require.Equal(t, "tool-1", msg.blocks[0].ID)
	require.Equal(t, ".", msg.blocks[0].Input["path"])

	msgID = "msg-2"
	events = []agent.StreamEvent{
		{Type: agent.EventMessageStart, MessageID: msgID, Role: "assistant"},
		{Type: agent.EventContentStart, MessageID: msgID, BlockType: "thinking"},
		{Type: agent.EventContentDelta, MessageID: msgID, BlockType: "thinking", Delta: "plan"},
		{Type: agent.EventContentEnd, MessageID: msgID, BlockType: "thinking"},
		{Type: agent.EventContentStart, MessageID: msgID, BlockType: "text"},
		{Type: agent.EventContentDelta, MessageID: msgID, BlockType: "text", Delta: "hello"},
		{Type: agent.EventContentEnd, MessageID: msgID, BlockType: "text"},
		{Type: agent.EventMessageEnd, MessageID: msgID, Role: "assistant"},
	}

	msg = nil
	ok = false
	for _, ev := range events {
		msg, ok = assembler.Apply(ev)
	}

	require.True(t, ok)
	require.NotNil(t, msg)
	require.Len(t, msg.blocks, 2)
	require.Equal(t, "thinking", msg.blocks[0].Type)
	require.Equal(t, "plan", msg.blocks[0].Text)
	require.Equal(t, "text", msg.blocks[1].Type)
	require.Equal(t, "hello", msg.blocks[1].Text)
}
