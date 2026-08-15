-- +goose Up
-- ALTER TABLE pending_emails
    MODIFY COLUMN to_user_id INT NULL,
    ADD COLUMN direct_email TINYINT(1) NOT NULL DEFAULT 0;