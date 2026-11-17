-- +goose Up
-- Create enums
CREATE TYPE source_type_enum AS ENUM ('rss', 'html', 'youtube', 'document');
CREATE TYPE fetch_interval_enum AS ENUM ('hourly', 'daily', 'weekly', 'monthly');

-- Rename feeds to sources
ALTER TABLE feeds RENAME TO sources;

-- Rename feed_follows to source_follows
ALTER TABLE feed_follows RENAME TO source_follows;

-- Rename columns
ALTER TABLE source_follows RENAME COLUMN feed_id TO source_id;
ALTER TABLE posts RENAME COLUMN feed_id TO source_id;

-- Add new columns to sources
ALTER TABLE sources 
    ADD COLUMN source_type source_type_enum NOT NULL DEFAULT 'rss',
    ADD COLUMN fetch_interval fetch_interval_enum NOT NULL DEFAULT 'hourly',
    ADD COLUMN auth_credentials JSONB;

-- +goose Down
-- Revert columns
ALTER TABLE sources 
    DROP COLUMN auth_credentials,
    DROP COLUMN fetch_interval,
    DROP COLUMN source_type;

-- Revert column names
ALTER TABLE posts RENAME COLUMN source_id TO feed_id;
ALTER TABLE source_follows RENAME COLUMN source_id TO feed_id;

-- Revert table names
ALTER TABLE source_follows RENAME TO feed_follows;
ALTER TABLE sources RENAME TO feeds;

-- Drop enums
DROP TYPE fetch_interval_enum;
DROP TYPE source_type_enum;
