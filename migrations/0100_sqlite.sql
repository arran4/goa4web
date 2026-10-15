-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications ADD COLUMN activity_key TEXT;
ALTER TABLE notifications ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;
CREATE UNIQUE INDEX notifications_recipient_activity_idx ON notifications (users_idusers, activity_key);
UPDATE schema_version SET version = 100;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS notifications_recipient_activity_idx;
ALTER TABLE notifications DROP COLUMN priority;
ALTER TABLE notifications DROP COLUMN activity_key;
UPDATE schema_version SET version = 99;
-- +goose StatementEnd
