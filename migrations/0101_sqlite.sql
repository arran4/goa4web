-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications ADD COLUMN resource_type TEXT;
ALTER TABLE notifications ADD COLUMN resource_id INTEGER;
ALTER TABLE notifications ADD COLUMN activity_position INTEGER;
CREATE INDEX notifications_recipient_resource_activity_idx ON notifications (users_idusers, resource_type, resource_id, activity_position);
UPDATE schema_version SET version = 101;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS notifications_recipient_resource_activity_idx;
ALTER TABLE notifications DROP COLUMN activity_position;
ALTER TABLE notifications DROP COLUMN resource_id;
ALTER TABLE notifications DROP COLUMN resource_type;
UPDATE schema_version SET version = 100;
-- +goose StatementEnd
