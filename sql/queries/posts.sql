-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, source_id, raw_content, ai_summary, embedding)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;
--

-- name: GetPostsForUser :many
SELECT posts.* FROM posts
JOIN source_follows ON source_follows.source_id = posts.source_id
WHERE source_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;
--

-- name: SearchPostsByVector :many
SELECT posts.* FROM posts
JOIN source_follows ON source_follows.source_id = posts.source_id
WHERE source_follows.user_id = $1
ORDER BY posts.embedding <=> $2
LIMIT $3;
--
