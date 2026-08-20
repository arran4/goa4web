-- +goose Up
ALTER TABLE faq ADD COLUMN priority INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS faq_priority_idx ON faq (priority);
UPDATE schema_version SET version = 74;