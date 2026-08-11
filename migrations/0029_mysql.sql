-- +goose Up
ALTER TABLE user_emails
    ADD COLUMN verification_expires_at datetime DEFAULT NULL AFTER last_verification_code;

ALTER TABLE pending_passwords
    MODIFY COLUMN passwd_algorithm tinytext NOT NULL;

-- Record upgrade to schema version 29
