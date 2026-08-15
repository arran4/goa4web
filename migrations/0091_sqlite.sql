-- +goose Up
ALTER TABLE user_passkeys
    ADD COLUMN name TEXT NOT NULL DEFAULT 'Passkey' AFTER user_id;

-- ALTER TABLE user_passkeys
    DROP COLUMN name;