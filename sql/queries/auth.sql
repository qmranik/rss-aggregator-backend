-- name: VerifyUsername :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1) AS exists;

-- name: GetUserByUsername :one
SELECT id, password_hash FROM users WHERE username = $1;

-- name: CreateSession :exec
INSERT INTO user_sessions (session_id, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, id)
VALUES ($1, $2);

-- name: GetSessionIDByRefreshToken :one
SELECT id FROM refresh_tokens WHERE token = $1;

-- name: GetUserUUIDBySessionID :one
SELECT user_id FROM user_sessions WHERE session_id = $1;

-- name: DeleteSession :exec
DELETE FROM user_sessions WHERE session_id = $1;
