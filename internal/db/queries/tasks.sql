-- name: CreateTask :one
INSERT INTO tasks (user_id, title, description, category_id, is_completed, due_date)
VALUES ($1, $2, $3, $4, COALESCE($5, FALSE), $6)
RETURNING id, user_id, title, description, category_id, is_completed, due_date, created_at, updated_at;

-- name: ListTasksByUser :many
SELECT id, user_id, title, description, category_id, is_completed, due_date, created_at, updated_at
FROM tasks
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CompleteTask :one
UPDATE tasks
SET is_completed = TRUE,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, title, description, category_id, is_completed, due_date, created_at, updated_at;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;
