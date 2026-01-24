-- name: CreateTask :exec
INSERT INTO tasks (id, repository_id, title, intent, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetTask :one
SELECT
    t.id,
    t.repository_id,
    t.title,
    t.intent,
    t.status,
    t.position,
    t.created_at,
    t.updated_at,
    r.full_name as repository_name
FROM tasks t
LEFT JOIN repositories r ON t.repository_id = r.id
WHERE t.id = ?;

-- name: ListTasks :many
SELECT * FROM tasks
ORDER BY status ASC, position ASC, created_at DESC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE status = ?
ORDER BY status ASC, position ASC, created_at DESC;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = ? WHERE id = ?;

-- name: UpdateTaskPosition :exec
UPDATE tasks SET position = ? WHERE id = ?;

-- name: UpdateTaskPositionAndStatus :exec
UPDATE tasks SET status = ?, position = ? WHERE id = ?;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;

-- name: ListTasksWithRepository :many
SELECT
    t.id,
    t.repository_id,
    t.title,
    t.intent,
    t.status,
    t.position,
    t.created_at,
    t.updated_at,
    r.full_name as repository_name,
    COALESCE((SELECT m.content FROM messages m WHERE m.task_id = t.id AND m.role = 'assistant' ORDER BY m.created_at DESC LIMIT 1), '') as last_assistant_message
FROM tasks t
LEFT JOIN repositories r ON t.repository_id = r.id
ORDER BY t.status ASC, t.position ASC, t.created_at DESC;
