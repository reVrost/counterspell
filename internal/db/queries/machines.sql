-- name: CreateMachine :one
INSERT INTO machines (id, name, mode, capabilities, created_at, updated_at, last_seen_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMachine :one
SELECT * FROM machines WHERE id = ? LIMIT 1;

-- name: UpdateMachineLastSeen :exec
UPDATE machines SET last_seen_at = ?, updated_at = ?
WHERE id = ?;

-- name: GetAllMachines :many
SELECT * FROM machines ORDER BY created_at DESC;

-- name: DeleteMachine :exec
DELETE FROM machines WHERE id = ?;
