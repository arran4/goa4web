-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications ADD COLUMN activity_key TEXT;
ALTER TABLE notifications ADD COLUMN priority INTEGER NOT NULL DEFAULT 0;
CREATE UNIQUE INDEX notifications_recipient_activity_idx ON notifications (users_idusers, activity_key);
CREATE TABLE IF NOT EXISTS comment_appends (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comment_id INTEGER NOT NULL,
    users_idusers INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS comment_appends_comment_idx ON comment_appends (comment_id);
UPDATE schema_version SET version = 102;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comment_appends;
DROP INDEX IF EXISTS notifications_recipient_activity_idx;
ALTER TABLE notifications DROP COLUMN priority;
ALTER TABLE notifications DROP COLUMN activity_key;
UPDATE schema_version SET version = 101;
-- +goose StatementEnd
