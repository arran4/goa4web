-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications
    ADD COLUMN resource_type VARCHAR(64) NULL,
    ADD COLUMN resource_id INT NULL,
    ADD COLUMN activity_position BIGINT NULL,
    ADD KEY notifications_recipient_resource_activity_idx (users_idusers, resource_type, resource_id, activity_position);
UPDATE schema_version SET version = 103;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE notifications
    DROP INDEX notifications_recipient_resource_activity_idx,
    DROP COLUMN activity_position,
    DROP COLUMN resource_id,
    DROP COLUMN resource_type;
UPDATE schema_version SET version = 102;
-- +goose StatementEnd
