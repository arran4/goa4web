-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications
    ADD COLUMN activity_key VARCHAR(255) NULL,
    ADD COLUMN priority TINYINT UNSIGNED NOT NULL DEFAULT 0,
    ADD UNIQUE KEY notifications_recipient_activity_idx (users_idusers, activity_key);
UPDATE schema_version SET version = 100;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE notifications
    DROP INDEX notifications_recipient_activity_idx,
    DROP COLUMN priority,
    DROP COLUMN activity_key;
UPDATE schema_version SET version = 99;
-- +goose StatementEnd
