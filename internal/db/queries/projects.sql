-- name: CreateProject :one
INSERT INTO projects (id, user_id, github_owner, github_repo, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, github_owner, github_repo) DO UPDATE SET github_owner = EXCLUDED.github_owner
RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE id = $1 AND user_id = $2;

-- name: GetProjects :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetRecentProjects :many
SELECT * FROM projects WHERE user_id = $1 ORDER BY created_at DESC LIMIT 5;

-- name: GetProjectByRepo :one
SELECT * FROM projects WHERE user_id = $1 AND github_repo = $2 LIMIT 1;

-- name: ProjectExists :one
SELECT EXISTS(SELECT 1 FROM projects WHERE user_id = $1 AND github_owner = $2 AND github_repo = $3);

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1 AND user_id = $2;

-- name: DeleteAllProjects :execresult
DELETE FROM projects WHERE user_id = $1;
