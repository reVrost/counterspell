-- name: GetWorkspace :one
SELECT * FROM workspaces WHERE id = ?;

-- name: ListWorkspaces :many
SELECT * FROM workspaces
ORDER BY name ASC;

-- name: CreateWorkspace :one
INSERT INTO workspaces (
    id, github_connection_id, name, local_path, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?
) RETURNING *;
