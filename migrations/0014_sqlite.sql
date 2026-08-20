-- +goose Up
ALTER TABLE pending_emails ADD COLUMN error_count INT NOT NULL DEFAULT 0;
UPDATE schema_version SET version = 14;