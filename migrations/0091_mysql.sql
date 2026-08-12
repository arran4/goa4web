-- +goose Up
ALTER TABLE user_passkeys
    ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT 'Passkey' AFTER user_id;

-- +goose Down
ALTER TABLE user_passkeys
    DROP COLUMN name;
