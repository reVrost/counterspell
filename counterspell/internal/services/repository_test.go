package services

import (
	"context"
	"testing"
	"time"

	"github.com/revrost/counterspell/internal/db/sqlc"
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

func TestGetLatestAgentRun_SessionID(t *testing.T) {
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

	// 2. Create first agent run with a backend_session_id
	run1ID, err := repo.CreateAgentRun(ctx, task.ID, "first prompt", "claude-code", "anthropic", "claude-3")
	require.NoError(t, err)

	sessionID1 := "session-uuid-1"
	err = repo.UpdateAgentRunBackendSessionID(ctx, run1ID, sessionID1)
	require.NoError(t, err)

	// 3. Verify GetLatestAgentRun returns the first run with its session_id
	latestRun, err := repo.GetLatestAgentRun(ctx, task.ID)
	require.NoError(t, err)
	require.NotNil(t, latestRun)
	assert.Equal(t, run1ID, latestRun.ID)
	assert.True(t, latestRun.BackendSessionID.Valid, "BackendSessionID should be valid")
	assert.Equal(t, sessionID1, latestRun.BackendSessionID.String, "Should have the correct session ID")

	// 4. Create a second agent run (without session_id initially)
	// Add a small delay to ensure different created_at timestamps
	time.Sleep(time.Millisecond * 2)
	run2ID, err := repo.CreateAgentRun(ctx, task.ID, "second prompt", "claude-code", "anthropic", "claude-3")
	require.NoError(t, err)

	// 5. Verify GetLatestAgentRun now returns the SECOND run
	latestRun, err = repo.GetLatestAgentRun(ctx, task.ID)
	require.NoError(t, err)
	require.NotNil(t, latestRun)
	assert.Equal(t, run2ID, latestRun.ID, "Should return the second run")
	assert.False(t, latestRun.BackendSessionID.Valid, "New run should not have session_id yet")

	// 6. Set session_id on second run
	sessionID2 := "session-uuid-2"
	err = repo.UpdateAgentRunBackendSessionID(ctx, run2ID, sessionID2)
	require.NoError(t, err)

	// 7. Verify GetLatestAgentRun returns the second run with its new session_id
	latestRun, err = repo.GetLatestAgentRun(ctx, task.ID)
	require.NoError(t, err)
	require.NotNil(t, latestRun)
	assert.Equal(t, run2ID, latestRun.ID)
	assert.True(t, latestRun.BackendSessionID.Valid, "BackendSessionID should be valid")
	assert.Equal(t, sessionID2, latestRun.BackendSessionID.String, "Should have the correct session ID")

	// 8. Verify first run still has its original session_id
	firstRun, err := repo.db.Queries.GetAgentRun(ctx, run1ID)
	require.NoError(t, err)
	assert.Equal(t, sessionID1, firstRun.BackendSessionID.String, "First run should still have its original session ID")
}
