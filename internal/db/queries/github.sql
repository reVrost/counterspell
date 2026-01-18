-- name: CreateGitHubConnection :exec
INSERT INTO github_connections (id, user_id, type, login, avatar_url, token, scope, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetActiveGitHubConnection :one
SELECT * FROM github_connections WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1;

-- name: DeleteAllGitHubConnections :execresult
DELETE FROM github_connections WHERE user_id = $1;
