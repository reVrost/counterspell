package services

import (
	"context"
	"testing"
	"time"

	"github.com/revrost/code/counterspell/internal/agent"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSaveMessage_PingPong tests that messages are stored correctly in a ping-pong fashion
func TestSaveMessage_PingPong(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	orch, err := NewOrchestrator(
		NewRepository(testDB),
		NewEventBus(),
		nil, nil, ":memory:",
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

	// User: HI
	userMsg1 := agent.Message{
		Role: "user",
		Content: []agent.ContentBlock{{Type: "text", Text: "HI"}},
	}
	orch.saveMessage(taskID, userMsg1)
	time.Sleep(2 * time.Millisecond)

	// Agent: HI there
	assistantMsg1 := agent.Message{
		Role: "assistant",
		Content: []agent.ContentBlock{{Type: "text", Text: "HI there"}},
	}
	orch.saveMessage(taskID, assistantMsg1)
	time.Sleep(2 * time.Millisecond)

	// User: how areyou
	userMsg2 := agent.Message{
		Role: "user",
		Content: []agent.ContentBlock{{Type: "text", Text: "how areyou"}},
	}
	orch.saveMessage(taskID, userMsg2)
	time.Sleep(2 * time.Millisecond)

	// Agent: good, and you?
	assistantMsg2 := agent.Message{
		Role: "assistant",
		Content: []agent.ContentBlock{{Type: "text", Text: "good, and you?"}},
	}
	orch.saveMessage(taskID, assistantMsg2)

	// Verify total 4 messages stored
	messages, err := orch.repo.GetMessagesByTask(context.Background(), taskID)
	require.NoError(t, err)
	assert.Equal(t, 4, len(messages), "Should have exactly 4 messages in DB")

	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "HI", messages[0].Content)
	assert.Equal(t, runID, messages[0].RunID)

	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, "HI there", messages[1].Content)

	assert.Equal(t, "user", messages[2].Role)
	assert.Equal(t, "how areyou", messages[2].Content)

	assert.Equal(t, "assistant", messages[3].Role)
	assert.Equal(t, "good, and you?", messages[3].Content)
}
