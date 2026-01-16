-- name: CreateAgentLog :exec
INSERT INTO agent_logs (task_id, level, message) VALUES (?, ?, ?);

-- name: GetAgentLogsByTask :many
SELECT * FROM agent_logs WHERE task_id = ? ORDER BY id ASC;
