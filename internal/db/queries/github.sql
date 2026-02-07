-- name: GetGithubConnection :one
SELECT * FROM github_connections LIMIT 1;

-- name: GetGithubConnectionByID :one
SELECT * FROM github_connections WHERE id = ?;

-- name: CreateGithubConnection :one
INSERT INTO github_connections (
    id, github_user_id, access_token, username, avatar_url, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: UpdateGithubConnection :one
UPDATE github_connections
SET access_token = ?, username = ?, avatar_url = ?, updated_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteGithubConnection :exec
DELETE FROM github_connections WHERE id = ?;
