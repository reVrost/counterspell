-- name: CreateConnectorOAuthAttempt :exec
INSERT INTO connector_oauth_attempts (connector, state, code_verifier, created_at)
VALUES (?, ?, ?, ?);

-- name: GetConnectorOAuthAttempt :one
SELECT connector, state, code_verifier, created_at
FROM connector_oauth_attempts
WHERE connector = ? AND state = ?;

-- name: DeleteConnectorOAuthAttempt :exec
DELETE FROM connector_oauth_attempts
WHERE connector = ? AND state = ?;

-- name: CleanupExpiredConnectorOAuthAttempts :exec
DELETE FROM connector_oauth_attempts
WHERE connector = ? AND created_at < ?;

-- name: GetConnectorAuth :one
SELECT connector, access_token, refresh_token, account_id, metadata_json, expires_at, connected_at, updated_at
FROM connector_auth
WHERE connector = ?;

-- name: UpsertConnectorAuth :exec
INSERT INTO connector_auth (
    connector,
    access_token,
    refresh_token,
    account_id,
    metadata_json,
    expires_at,
    connected_at,
    updated_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
)
ON CONFLICT(connector) DO UPDATE SET
    access_token = excluded.access_token,
    refresh_token = excluded.refresh_token,
    account_id = excluded.account_id,
    metadata_json = excluded.metadata_json,
    expires_at = excluded.expires_at,
    connected_at = excluded.connected_at,
    updated_at = excluded.updated_at;

-- name: DeleteConnectorAuth :exec
DELETE FROM connector_auth
WHERE connector = ?;
