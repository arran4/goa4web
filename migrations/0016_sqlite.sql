-- +goose Up
ALTER TABLE pending_emails ADD COLUMN to_user_id INT NOT NULL DEFAULT 0;
UPDATE schema_version SET version = 16;