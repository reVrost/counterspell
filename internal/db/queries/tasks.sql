-- name: CreateTask :one
INSERT INTO tasks (id, user_id, project_id, title, intent, status, position, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1 AND user_id = $2;

-- name: ListTasks :many
SELECT * FROM tasks
WHERE user_id = $1
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE user_id = $1 AND status = $2
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByProject :many
SELECT * FROM tasks
WHERE user_id = $1 AND project_id = $2
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatusAndProject :many
SELECT * FROM tasks
WHERE user_id = $1 AND status = $2 AND project_id = $3
ORDER BY status ASC, position ASC, created_at DESC;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = $1, updated_at = $2 WHERE id = $3 AND user_id = $4;

-- name: UpdateTaskPosition :exec
UPDATE tasks SET position = $1, updated_at = $2 WHERE id = $3 AND user_id = $4;

-- name: UpdateTaskPositionAndStatus :exec
UPDATE tasks SET status = $1, position = $2, updated_at = $3 WHERE id = $4 AND user_id = $5;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1 AND user_id = $2;

-- name: UpdateTaskResult :exec
UPDATE tasks SET status = $1, agent_output = $2, git_diff = $3, message_history = $4, updated_at = $5 
WHERE id = $6 AND user_id = $7;

-- name: ClearTaskHistory :exec
UPDATE tasks SET message_history = '', agent_output = '', updated_at = $1 WHERE id = $2 AND user_id = $3;

-- name: ResetCompletedInProgressTasks :execresult
UPDATE tasks SET status = $1, updated_at = $2
WHERE user_id = $3 AND status = $4 AND (agent_output IS NOT NULL AND agent_output != '');

-- name: ResetStuckInProgressTasks :execresult
UPDATE tasks SET status = $1, updated_at = $2
WHERE user_id = $3 AND status = $4 AND (agent_output IS NULL OR agent_output = '');
