-- name: CreateOAuthLoginAttempt :exec
INSERT INTO oauth_login_attempts (state, code_verifier, created_at)
VALUES (?, ?, ?);

-- name: GetOAuthLoginAttempt :one
SELECT state, code_verifier, created_at
FROM oauth_login_attempts
WHERE state = ?;

-- name: DeleteOAuthLoginAttempt :exec
DELETE FROM oauth_login_attempts WHERE state = ?;

-- name: CleanupExpiredOAuthAttempts :exec
DELETE FROM oauth_login_attempts WHERE created_at < ?;

-- name: CreateMachineIdentity :exec
INSERT INTO machine_identity (machine_id, machine_jwt, user_id, subdomain, tunnel_provider, tunnel_token, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetMachineIdentity :one
SELECT machine_id, machine_jwt, user_id, subdomain, tunnel_provider, tunnel_token, created_at, last_seen_at
FROM machine_identity
WHERE machine_id = ?;

-- name: UpdateMachineIdentityLastSeen :exec
UPDATE machine_identity SET last_seen_at = ? WHERE machine_id = ?;

-- name: UpdateMachineIdentityJWT :exec
UPDATE machine_identity SET machine_jwt = ? WHERE machine_id = ?;

-- name: UpsertMachineIdentity :one
INSERT INTO machine_identity (machine_id, machine_jwt, user_id, subdomain, tunnel_provider, tunnel_token, created_at, last_seen_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(machine_id) DO UPDATE SET
    machine_jwt = excluded.machine_jwt,
    user_id = excluded.user_id,
    subdomain = excluded.subdomain,
    tunnel_provider = excluded.tunnel_provider,
    tunnel_token = excluded.tunnel_token,
    last_seen_at = excluded.last_seen_at
RETURNING *;

-- name: GetMachineByUserID :one
SELECT machine_id, machine_jwt, user_id, subdomain, tunnel_provider, tunnel_token, created_at, last_seen_at
FROM machine_identity
WHERE user_id = ?;
