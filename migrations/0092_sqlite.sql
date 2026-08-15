-- +goose Up
ALTER TABLE user_passkeys
    ADD COLUMN backup_eligible BOOLEAN DEFAULT NULL AFTER name,
    ADD COLUMN backup_state BOOLEAN DEFAULT NULL AFTER backup_eligible;

-- ALTER TABLE user_passkeys
    DROP COLUMN backup_state,
    DROP COLUMN backup_eligible;