-- name: CreateUser :one
INSERT INTO users (user_name, email, password_hash, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING id, user_name, role, email, avatar_url, created_at;
-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE ID = $1 LIMIT 1;
-- name: IncrementTokenVersion :one
UPDATE users
SET token_version = token_version + 1
WHERE id = $1
RETURNING token_version;
