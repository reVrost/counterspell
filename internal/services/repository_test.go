package services

import (
	"context"
	"testing"
	"time"

	"github.com/revrost/code/counterspell/internal/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListWithRepository_LastAssistantMessage(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewRepository(testDB)
	ctx := context.Background()

	// 1. Setup repo and task
	conn, err := repo.db.Queries.CreateGithubConnection(ctx, sqlc.CreateGithubConnectionParams{
		ID:           "conn-1",
		GithubUserID: "user-1",
		AccessToken:  "token",
		Username:     "testuser",
	})
	require.NoError(t, err)

	repoRow, err := repo.db.Queries.CreateRepository(ctx, sqlc.CreateRepositoryParams{
		ID:           "repo-1",
		ConnectionID: conn.ID,
		Name:         "test-repo",
		FullName:     "test/test-repo",
		Owner:        "test",
	})
	require.NoError(t, err)

	task, err := repo.Create(ctx, repoRow.ID, "test task")
	require.NoError(t, err)

	// 2. Test fetching with NO messages (should not error, should be nil)
	tasks, err := repo.ListWithRepository(ctx)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, task.ID, tasks[0].ID)
	assert.Nil(t, tasks[0].LastAssistantMessage, "LastAssistantMessage should be nil when no messages exist")

	// 3. Add a user message only (still should be nil)
	runID, err := repo.CreateAgentRun(ctx, task.ID, "prompt", "native", "anthropic", "model")
	require.NoError(t, err)
	err = repo.CreateMessage(ctx, task.ID, runID, "user", "hello")
	require.NoError(t, err)

	tasks, err = repo.ListWithRepository(ctx)
	require.NoError(t, err)
	assert.Nil(t, tasks[0].LastAssistantMessage, "LastAssistantMessage should be nil when only user messages exist")

	// 4. Add an assistant message
	err = repo.CreateMessage(ctx, task.ID, runID, "assistant", "I am here to help")
	require.NoError(t, err)

	time.Sleep(time.Millisecond)

	tasks, err = repo.ListWithRepository(ctx)
	require.NoError(t, err)
	require.NotNil(t, tasks[0].LastAssistantMessage)
	assert.Equal(t, "I am here to help", *tasks[0].LastAssistantMessage)

	// 5. Add another assistant message (should get the latest)
	err = repo.CreateMessage(ctx, task.ID, runID, "assistant", "Second message")
	require.NoError(t, err)

	time.Sleep(time.Millisecond)

	tasks, err = repo.ListWithRepository(ctx)
	require.NoError(t, err)
	require.NotNil(t, tasks[0].LastAssistantMessage)
	assert.Equal(t, "Second message", *tasks[0].LastAssistantMessage)
}
