-- name: GetSourceFollowsForUser :many
SELECT * FROM source_follows WHERE user_id = $1;
--

-- name: CreateSourceFollow :one
INSERT INTO source_follows (id, created_at, updated_at, user_id, source_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
--

-- name: DeleteSourceFollow :exec
DELETE FROM source_follows WHERE id = $1 and user_id = $2;
--
