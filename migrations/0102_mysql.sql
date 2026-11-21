-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications
    ADD COLUMN activity_key VARCHAR(255) NULL,
    ADD COLUMN priority TINYINT UNSIGNED NOT NULL DEFAULT 0,
    ADD UNIQUE KEY notifications_recipient_activity_idx (users_idusers, activity_key);
CREATE TABLE IF NOT EXISTS comment_appends (
    id INT NOT NULL AUTO_INCREMENT,
    comment_id INT NOT NULL,
    users_idusers INT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY comment_appends_comment_idx (comment_id)
);
UPDATE schema_version SET version = 102;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comment_appends;
ALTER TABLE notifications
    DROP INDEX notifications_recipient_activity_idx,
    DROP COLUMN priority,
    DROP COLUMN activity_key;
UPDATE schema_version SET version = 101;
-- +goose StatementEnd
