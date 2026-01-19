-- name: CreateMachine :exec
INSERT INTO machines (id, name, mode, capabilities, last_seen_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetMachine :one
SELECT * FROM machines WHERE id = ?;

-- name: ListMachines :many
SELECT * FROM machines
ORDER BY name ASC;

-- name: ListMachinesByMode :many
SELECT * FROM machines
WHERE mode = ?
ORDER BY name ASC;

-- name: UpdateMachineLastSeen :exec
UPDATE machines SET last_seen_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: UpdateMachineCapabilities :exec
UPDATE machines SET capabilities = ? WHERE id = ?;

-- name: DeleteMachine :exec
DELETE FROM machines WHERE id = ?;

-- name: GetDefaultMachine :one
SELECT * FROM machines
WHERE mode = 'local'
LIMIT 1;
