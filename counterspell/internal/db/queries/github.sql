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

-- name: ListRepositories :many
SELECT * FROM repositories WHERE connection_id = ? ORDER BY full_name ASC;

-- name: GetRepository :one
SELECT * FROM repositories WHERE id = ?;

-- name: CreateRepository :one
INSERT INTO repositories (
    id, connection_id, name, full_name, owner, is_private, html_url, clone_url, local_path, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: UpsertRepository :one
INSERT INTO repositories (
    id, connection_id, name, full_name, owner, is_private, html_url, clone_url, local_path, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT(connection_id, full_name) DO UPDATE SET
    name = excluded.name,
    is_private = excluded.is_private,
    html_url = excluded.html_url,
    clone_url = excluded.clone_url,
    local_path = excluded.local_path,
    updated_at = excluded.updated_at
RETURNING *;

-- name: DeleteRepositoriesByConnection :exec
DELETE FROM repositories WHERE connection_id = ?;
