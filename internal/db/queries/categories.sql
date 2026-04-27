-- name: CreateCategory :one
INSERT INTO categories (name, color)
VALUES ($1, $2)
RETURNING id, name, color, created_at, updated_at;

-- name: ListCategories :many
SELECT id, name, color, created_at, updated_at
FROM categories
ORDER BY name ASC;

-- name: GetCategoryByID :one
SELECT id, name, color, created_at, updated_at
FROM categories
WHERE id = $1;
