-- name: GetWorkspace :one
SELECT * FROM workspaces WHERE id = ?;

-- name: ListWorkspaces :many
SELECT * FROM workspaces
ORDER BY name ASC;
