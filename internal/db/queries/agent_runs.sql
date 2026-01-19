-- name: CreateAgentRun :exec
INSERT INTO agent_runs (id, task_id, agent_backend, prompt, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetAgentRun :one
SELECT * FROM agent_runs WHERE id = ?;

-- name: ListAgentRunsByTask :many
SELECT * FROM agent_runs
WHERE task_id = ?
ORDER BY created_at ASC;

-- name: GetLatestRun :one
SELECT * FROM agent_runs
WHERE task_id = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteAgentRunsByTask :exec
DELETE FROM agent_runs WHERE task_id = ?;
