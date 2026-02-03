package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/revrost/counterspell/internal/agent"
	"github.com/revrost/counterspell/internal/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConsumeAgentStream_PersistsMessages verifies stream events persist final message parts.
func TestConsumeAgentStream_PersistsMessages(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	orch, err := NewOrchestrator(
		NewRepository(testDB),
		NewEventBus(),
		nil, nil,
		stubRepoManager{},
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Create github connection, repository and task to satisfy FK constraints
	conn, err := orch.repo.db.Queries.CreateGithubConnection(ctx, sqlc.CreateGithubConnectionParams{
		ID:           "conn-1",
		GithubUserID: "user-1",
		AccessToken:  "token",
		Username:     "testuser",
	})
	require.NoError(t, err)

	repoRow, err := orch.repo.db.Queries.CreateRepository(ctx, sqlc.CreateRepositoryParams{
		ID:           "repo-1",
		ConnectionID: conn.ID,
		Name:         "test-repo",
		FullName:     "test/test-repo",
		Owner:        "test",
	})
	require.NoError(t, err)

	task, err := orch.repo.Create(ctx, repoRow.ID, "start")
	require.NoError(t, err)
	taskID := task.ID

	// Create agent run
	runID, err := orch.repo.CreateAgentRun(ctx, taskID, "start", "native", "anthropic", "claude-3")
	require.NoError(t, err)

	events := make(chan agent.StreamEvent, 32)
	stream := &agent.Stream{Events: events}

	go func() {
		defer close(events)

		// Assistant message with thinking + text
		events <- agent.StreamEvent{Type: agent.EventMessageStart, MessageID: "msg-1", Role: "assistant"}
		events <- agent.StreamEvent{Type: agent.EventContentStart, MessageID: "msg-1", BlockType: "thinking"}
		events <- agent.StreamEvent{Type: agent.EventContentDelta, MessageID: "msg-1", BlockType: "thinking", Delta: "planning..."}
		events <- agent.StreamEvent{Type: agent.EventContentEnd, MessageID: "msg-1", BlockType: "thinking"}
		events <- agent.StreamEvent{Type: agent.EventContentStart, MessageID: "msg-1", BlockType: "text"}
		events <- agent.StreamEvent{Type: agent.EventContentDelta, MessageID: "msg-1", BlockType: "text", Delta: "HI there"}
		events <- agent.StreamEvent{Type: agent.EventContentEnd, MessageID: "msg-1", BlockType: "text"}
		events <- agent.StreamEvent{Type: agent.EventMessageEnd, MessageID: "msg-1", Role: "assistant"}

		// Tool use
		events <- agent.StreamEvent{Type: agent.EventMessageStart, MessageID: "msg-2", Role: "assistant"}
		events <- agent.StreamEvent{
			Type:      agent.EventContentStart,
			MessageID: "msg-2",
			BlockType: "tool_use",
			Block:     &agent.ContentBlock{Type: "tool_use", Name: "ls", ID: "tool-1"},
		}
		events <- agent.StreamEvent{Type: agent.EventContentDelta, MessageID: "msg-2", BlockType: "tool_use", Delta: `{"path":"."}`}
		events <- agent.StreamEvent{Type: agent.EventContentEnd, MessageID: "msg-2", BlockType: "tool_use"}
		events <- agent.StreamEvent{Type: agent.EventMessageEnd, MessageID: "msg-2", Role: "assistant"}

		// Tool result
		events <- agent.StreamEvent{Type: agent.EventMessageStart, MessageID: "msg-3", Role: "user"}
		events <- agent.StreamEvent{
			Type:      agent.EventContentStart,
			MessageID: "msg-3",
			BlockType: "tool_result",
			Block:     &agent.ContentBlock{Type: "tool_result", ToolUseID: "tool-1", Content: "file1.txt"},
		}
		events <- agent.StreamEvent{Type: agent.EventContentEnd, MessageID: "msg-3", BlockType: "tool_result"}
		events <- agent.StreamEvent{Type: agent.EventMessageEnd, MessageID: "msg-3", Role: "user"}
	}()

	err = orch.consumeAgentStream(ctx, taskID, runID, stream)
	require.NoError(t, err)

	// Verify total 4 messages stored
	messages, err := orch.repo.GetMessagesByTask(context.Background(), taskID)
	require.NoError(t, err)
	require.Len(t, messages, 3)

	var assistantMsg *sqlc.Message
	var toolUseMsg *sqlc.Message
	var toolResultMsg *sqlc.Message
	for i := range messages {
		switch messages[i].Content {
		case "HI there":
			assistantMsg = &messages[i]
		case "ls":
			toolUseMsg = &messages[i]
		case "file1.txt":
			toolResultMsg = &messages[i]
		}
	}

	require.NotNil(t, assistantMsg)
	require.NotNil(t, toolUseMsg)
	require.NotNil(t, toolResultMsg)

	assert.Equal(t, "assistant", assistantMsg.Role)
	assert.Equal(t, runID, assistantMsg.RunID)

	var parts []agent.ContentBlock
	require.NoError(t, json.Unmarshal([]byte(assistantMsg.Parts), &parts))
	require.Len(t, parts, 2)
	assert.Equal(t, "thinking", parts[0].Type)
	assert.Equal(t, "planning...", parts[0].Text)
	assert.Equal(t, "text", parts[1].Type)

	assert.Equal(t, "tool", toolUseMsg.Role)
	assert.Equal(t, "tool", toolResultMsg.Role)
}

func TestContinueTask_Validation(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	orch, err := NewOrchestrator(
		NewRepository(testDB),
		NewEventBus(),
		nil, nil,
		stubRepoManager{},
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Test empty message
	err = orch.ContinueTask(ctx, "task-1", "", "model-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// Test task not found
	err = orch.ContinueTask(ctx, "non-existent", "hi", "model-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
}

func TestExecuteTask_BackendSelection(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewRepository(testDB)
	settingsSvc := NewSettingsService(testDB)
	orch, err := NewOrchestrator(
		repo,
		NewEventBus(),
		settingsSvc,
		nil,
		stubRepoManager{},
	)
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Test default backend (native)
	task1, err := repo.Create(ctx, "", "test intent 1")
	require.NoError(t, err)

	job1 := TaskJob{
		TaskID:   task1.ID,
		Intent:   "test intent 1",
		ResultCh: make(chan TaskResult, 1),
	}

	// We can't easily intercept the backend instantiation without refactoring more,
	// but we can check what was saved in the agent_runs table.
	orch.executeTask(ctx, job1)

	run1, err := repo.GetLatestAgentRun(ctx, task1.ID)
	require.NoError(t, err)
	assert.Equal(t, "native", run1.AgentBackend, "Should default to native backend")

	// 2. Test explicit claude-code backend
	err = settingsSvc.UpdateSettings(ctx, &Settings{
		AgentBackend: "claude-code",
		AnthropicKey: "test-key",
	})
	require.NoError(t, err)

	task2, err := repo.Create(ctx, "", "test intent 2")
	require.NoError(t, err)

	job2 := TaskJob{
		TaskID:   task2.ID,
		Intent:   "test intent 2",
		ResultCh: make(chan TaskResult, 1),
	}

	// This will likely fail to actually RUN because 'claude' binary isn't there,
	// but it should still CREATE the agent run with the correct backend type.
	orch.executeTask(ctx, job2)

	run2, err := repo.GetLatestAgentRun(ctx, task2.ID)
	require.NoError(t, err)
	assert.Equal(t, "claude-code", run2.AgentBackend, "Should use claude-code backend from settings")
}

// TestContinueTask_WithMessageHistory tests that continue task loads and passes message history
func TestContinueTask_WithMessageHistory(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewRepository(testDB)
	orch, err := NewOrchestrator(
		repo,
		NewEventBus(),
		nil, nil, stubRepoManager{},
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Create task with some message history
	conn, err := orch.repo.db.Queries.CreateGithubConnection(ctx, sqlc.CreateGithubConnectionParams{
		ID:           "conn-1",
		GithubUserID: "user-1",
		AccessToken:  "token",
		Username:     "testuser",
	})
	require.NoError(t, err)

	repoRow, err := orch.repo.db.Queries.CreateRepository(ctx, sqlc.CreateRepositoryParams{
		ID:           "repo-1",
		ConnectionID: conn.ID,
		Name:         "test-repo",
		FullName:     "test/test-repo",
		Owner:        "test",
	})
	require.NoError(t, err)

	task, err := repo.Create(ctx, repoRow.ID, "initial intent")
	require.NoError(t, err)

	runID, err := repo.CreateAgentRun(ctx, task.ID, "initial intent", "native", "anthropic", "claude-3")
	require.NoError(t, err)

	// Save some messages to DB
	err = repo.CreateMessage(ctx, task.ID, runID, "user", "Hello, I need help")
	require.NoError(t, err)
	err = repo.CreateMessage(ctx, task.ID, runID, "assistant", "Sure, how can I help?")
	require.NoError(t, err)
	err = repo.CreateMessage(ctx, task.ID, runID, "user", "Can you write a function?")
	require.NoError(t, err)

	// Verify messages are in DB
	messages, err := repo.GetMessagesByTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, len(messages), "Should have 3 messages")

	// Test ConvertMessagesToJSON
	jsonStr, err := ConvertMessagesToJSON(messages)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr, "JSON should not be empty")
	assert.Contains(t, jsonStr, "Hello, I need help", "JSON should contain user message")
	assert.Contains(t, jsonStr, "Sure, how can I help?", "JSON should contain assistant message")

	// Verify JSON structure is correct
	assert.Contains(t, jsonStr, "user", "Should have user role")
	assert.Contains(t, jsonStr, "assistant", "Should have assistant role")
	assert.Contains(t, jsonStr, "text", "Should have text content block")

	// Test that submitTaskJob loads message history correctly
	// We test this by verifying the job is submitted (async, so no immediate error)
	// The actual execution will fail due to missing git/API, but loading should work
	err = orch.submitTaskJob(ctx, task.ID, repoRow.ID, "continue", "model-1", "test", "test", "", true)
	require.NoError(t, err, "submitTaskJob should successfully load message history")
}

// TestContinueTask_NoMessages tests continuation with no message history (first continuation)
func TestContinueTask_NoMessages(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewRepository(testDB)
	orch, err := NewOrchestrator(
		repo,
		NewEventBus(),
		nil, nil, stubRepoManager{},
	)
	require.NoError(t, err)

	ctx := context.Background()

	// Create task but no messages
	conn, err := orch.repo.db.Queries.CreateGithubConnection(ctx, sqlc.CreateGithubConnectionParams{
		ID:           "conn-1",
		GithubUserID: "user-1",
		AccessToken:  "token",
		Username:     "testuser",
	})
	require.NoError(t, err)

	repoRow, err := orch.repo.db.Queries.CreateRepository(ctx, sqlc.CreateRepositoryParams{
		ID:           "repo-1",
		ConnectionID: conn.ID,
		Name:         "test-repo",
		FullName:     "test/test-repo",
		Owner:        "test",
	})
	require.NoError(t, err)

	task, err := repo.Create(ctx, repoRow.ID, "start")
	require.NoError(t, err)

	// Verify no messages
	messages, err := repo.GetMessagesByTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, len(messages), "Should have no messages initially")

	// Test ConvertMessagesToJSON with empty messages
	jsonStr, err := ConvertMessagesToJSON(messages)
	require.NoError(t, err)
	assert.Equal(t, "[]", jsonStr, "Empty messages should produce empty JSON array")

	// ContinueTask should work fine even with no message history
	// It will fail during execution (no git, no API keys) but the message history loading should succeed
	err = orch.submitTaskJob(ctx, task.ID, repoRow.ID, "continue", "model-1", "test", "test", "", true)
	require.NoError(t, err, "submitTaskJob should work even with no message history")
}
