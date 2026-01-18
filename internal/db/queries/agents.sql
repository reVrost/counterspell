-- name: CreateAgent :one
INSERT INTO agents (id, user_id, name, system_prompt, tools, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAgent :one
SELECT * FROM agents WHERE id = $1;

-- name: GetAgentByName :one
SELECT * FROM agents WHERE name = $1;

-- name: ListAgents :many
SELECT * FROM agents
ORDER BY name ASC;

-- name: ListAgentsByUser :many
SELECT * FROM agents
WHERE user_id = $1
ORDER BY name ASC;

-- name: UpdateAgent :exec
UPDATE agents 
SET name = $1, system_prompt = $2, tools = $3, updated_at = $4 
WHERE id = $5;

-- name: DeleteAgent :exec
DELETE FROM agents WHERE id = $1;
