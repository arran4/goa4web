-- +goose Up
ALTER TABLE pending_emails
    ADD COLUMN IF NOT EXISTS error_count INT NOT NULL DEFAULT 0;