-- name: CreateGitHubConnection :exec
INSERT INTO github_connections (id, type, login, avatar_url, token, scope, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetActiveGitHubConnection :one
SELECT * FROM github_connections ORDER BY created_at DESC LIMIT 1;

-- name: DeleteAllGitHubConnections :execresult
DELETE FROM github_connections;
