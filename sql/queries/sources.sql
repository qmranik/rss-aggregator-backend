-- name: CreateSource :one
INSERT INTO sources (id, created_at, updated_at, name, url, user_id, source_type, fetch_interval, auth_credentials)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetSources :many
SELECT * FROM sources;

-- name: GetNextSourcesToFetch :many
SELECT * FROM sources
WHERE last_fetched_at IS NULL
   OR last_fetched_at < NOW() - (
        CASE fetch_interval
            WHEN 'hourly'  THEN INTERVAL '1 hour'
            WHEN 'daily'   THEN INTERVAL '1 day'
            WHEN 'weekly'  THEN INTERVAL '7 days'
            WHEN 'monthly' THEN INTERVAL '30 days'
        END
      )
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT $1;

-- name: MarkSourceFetched :one
UPDATE sources
SET last_fetched_at = NOW(),
updated_at = NOW()
WHERE id = $1
RETURNING *;
