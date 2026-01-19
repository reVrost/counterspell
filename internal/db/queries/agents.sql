-- name: CreateAgent :exec
INSERT INTO agents (id, name, system_prompt, tools, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetAgent :one
SELECT * FROM agents WHERE id = ?;

-- name: GetAgentByName :one
SELECT * FROM agents WHERE name = ?;

-- name: ListAgents :many
SELECT * FROM agents
ORDER BY name ASC;

-- name: UpdateAgent :exec
UPDATE agents
SET name = ?, system_prompt = ?, tools = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteAgent :exec
DELETE FROM agents WHERE id = ?;
