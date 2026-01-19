-- name: CreateAgentRun :exec
INSERT INTO agent_runs (id, task_id, step, agent_id, status, input, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetAgentRun :one
SELECT * FROM agent_runs WHERE id = ?;

-- name: ListAgentRunsByTask :many
SELECT * FROM agent_runs
WHERE task_id = ?
ORDER BY created_at ASC;

-- name: ListAgentRunsByTaskAndStep :many
SELECT * FROM agent_runs
WHERE task_id = ? AND step = ?
ORDER BY created_at ASC;

-- name: UpdateAgentRunStatus :exec
UPDATE agent_runs
SET status = ?, started_at = ?
WHERE id = ?;

-- name: CompleteAgentRun :exec
UPDATE agent_runs
SET status = 'completed', output = ?, message_history = ?, artifact_path = ?, completed_at = ?
WHERE id = ?;

-- name: FailAgentRun :exec
UPDATE agent_runs
SET status = 'failed', error = ?, message_history = ?, completed_at = ?
WHERE id = ?;

-- name: GetLatestRunForStep :one
SELECT * FROM agent_runs
WHERE task_id = ? AND step = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteAgentRunsByTask :exec
DELETE FROM agent_runs WHERE task_id = ?;
