-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, password_hash)
VALUES ( $1, NOW(), NOW(),$2, $3 )
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;
