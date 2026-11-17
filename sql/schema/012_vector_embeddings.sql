-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE posts ADD COLUMN embedding vector(768);
CREATE INDEX ON posts USING hnsw (embedding vector_cosine_ops);

-- +goose Down
DROP INDEX IF EXISTS posts_embedding_idx;
ALTER TABLE posts DROP COLUMN embedding;
DROP EXTENSION IF EXISTS vector;
