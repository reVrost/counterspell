-- name: CreateAgentRun :one
INSERT INTO agent_runs (id, task_id, step, agent_id, status, input, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAgentRun :one
SELECT * FROM agent_runs WHERE id = $1;

-- name: ListAgentRunsByTask :many
SELECT * FROM agent_runs
WHERE task_id = $1
ORDER BY created_at ASC;

-- name: ListAgentRunsByTaskAndStep :many
SELECT * FROM agent_runs
WHERE task_id = $1 AND step = $2
ORDER BY created_at ASC;

-- name: UpdateAgentRunStatus :exec
UPDATE agent_runs 
SET status = $1, started_at = $2
WHERE id = $3;

-- name: CompleteAgentRun :exec
UPDATE agent_runs 
SET status = 'completed', output = $1, message_history = $2, artifact_path = $3, completed_at = $4
WHERE id = $5;

-- name: FailAgentRun :exec
UPDATE agent_runs 
SET status = 'failed', error = $1, message_history = $2, completed_at = $3
WHERE id = $4;

-- name: AppendMessageHistory :exec
UPDATE agent_runs 
SET message_history = message_history || $1::jsonb
WHERE id = $2;

-- name: GetLatestRunForStep :one
SELECT * FROM agent_runs
WHERE task_id = $1 AND step = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteAgentRunsByTask :exec
DELETE FROM agent_runs WHERE task_id = $1;
