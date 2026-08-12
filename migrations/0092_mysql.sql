-- +goose Up
ALTER TABLE user_passkeys
    ADD COLUMN backup_eligible BOOLEAN DEFAULT NULL AFTER name,
    ADD COLUMN backup_state BOOLEAN DEFAULT NULL AFTER backup_eligible;

-- Existing rows intentionally remain NULL. Their eligibility is established
-- from the first successfully validated assertion instead of being assumed false.

-- +goose Down
ALTER TABLE user_passkeys
    DROP COLUMN backup_state,
    DROP COLUMN backup_eligible;
