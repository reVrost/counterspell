-- name: CreateAuth :one
INSERT INTO auth (machine_id, jwt_token, user_id, email, expires_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetAuth :one
SELECT * FROM auth WHERE machine_id = ? ORDER BY created_at DESC LIMIT 1;

-- name: GetAuthByMachineID :one
SELECT * FROM auth WHERE machine_id = ? LIMIT 1;

-- name: UpdateAuth :exec
UPDATE auth SET jwt_token = ?, expires_at = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteAuth :exec
DELETE FROM auth WHERE id = ?;

-- name: DeleteAuthByMachine :exec
DELETE FROM auth WHERE machine_id = ?;
