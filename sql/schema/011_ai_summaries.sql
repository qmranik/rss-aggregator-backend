-- +goose Up
ALTER TABLE posts ADD COLUMN ai_summary TEXT;
ALTER TABLE posts ADD COLUMN raw_content TEXT;

-- +goose Down
ALTER TABLE posts DROP COLUMN ai_summary;
ALTER TABLE posts DROP COLUMN raw_content;
