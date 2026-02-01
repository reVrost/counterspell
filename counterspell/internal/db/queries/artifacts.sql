-- name: CreateArtifact :exec
INSERT INTO artifacts (id, run_id, path, content, version, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetArtifact :one
SELECT * FROM artifacts WHERE id = ?;

-- name: GetArtifactsByRun :many
SELECT * FROM artifacts WHERE run_id = ? ORDER BY created_at ASC;

-- name: GetArtifactsByTask :many
SELECT a.* FROM artifacts a
JOIN agent_runs ar ON a.run_id = ar.id
WHERE ar.task_id = ?
ORDER BY a.created_at ASC;

-- name: DeleteArtifactsByRun :exec
DELETE FROM artifacts WHERE run_id = ?;
