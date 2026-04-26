-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, uuid, name, email, password_hash, streaks, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, uuid, name, email, password_hash, streaks, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, uuid, name, email, password_hash, streaks, created_at, updated_at
FROM users
WHERE email = $1;

-- name: UpdateUserStreaks :one
UPDATE users
SET streaks = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, uuid, name, email, password_hash, streaks, created_at, updated_at;
