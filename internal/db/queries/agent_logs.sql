-- name: CreateAgentLog :exec
INSERT INTO agent_logs (task_id, level, message) VALUES ($1, $2, $3);

-- name: GetAgentLogsByTask :many
SELECT * FROM agent_logs WHERE task_id = $1 ORDER BY id ASC;
