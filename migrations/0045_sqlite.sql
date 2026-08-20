-- +goose Up
ALTER TABLE pending_emails ADD COLUMN direct_email TINYINT(1) NOT NULL DEFAULT 0;
UPDATE schema_version SET version = 45;