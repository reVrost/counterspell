-- name: CreateProject :exec
INSERT INTO projects (id, github_owner, github_repo, created_at)
VALUES (?, ?, ?, ?);

-- name: GetProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: GetRecentProjects :many
SELECT * FROM projects ORDER BY created_at DESC LIMIT 5;

-- name: GetProjectByRepo :one
SELECT * FROM projects WHERE github_repo = ? LIMIT 1;

-- name: ProjectExists :one
SELECT EXISTS(SELECT 1 FROM projects WHERE github_owner = ? AND github_repo = ?);

-- name: DeleteAllProjects :execresult
DELETE FROM projects;
