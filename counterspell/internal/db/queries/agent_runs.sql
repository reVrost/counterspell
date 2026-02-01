-- name: CreateAgentRun :exec
INSERT INTO agent_runs (id, task_id, prompt, agent_backend, provider, model, backend_session_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateAgentRunCompleted :exec
UPDATE agent_runs SET completed_at = ? WHERE id = ?;

-- name: UpdateAgentRunBackendSessionID :exec
UPDATE agent_runs SET backend_session_id = ? WHERE id = ?;

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
