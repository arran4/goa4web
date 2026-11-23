-- +goose Up
-- +goose StatementBegin
CREATE TABLE digest_deliveries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INT NOT NULL,
    dismiss_notifications INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at DATETIME DEFAULT NULL
);
CREATE INDEX digest_deliveries_user_idx ON digest_deliveries (user_id);

CREATE TABLE digest_delivery_notifications (
    digest_delivery_id INT NOT NULL,
    notification_id INT NOT NULL,
    PRIMARY KEY (digest_delivery_id, notification_id)
);
CREATE INDEX digest_delivery_notifications_notification_idx ON digest_delivery_notifications (notification_id);

ALTER TABLE pending_emails ADD COLUMN digest_delivery_id INT DEFAULT NULL;
CREATE UNIQUE INDEX pending_emails_digest_delivery_idx ON pending_emails (digest_delivery_id);
UPDATE schema_version SET version = 101;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX pending_emails_digest_delivery_idx;
ALTER TABLE pending_emails DROP COLUMN digest_delivery_id;
DROP TABLE digest_delivery_notifications;
DROP TABLE digest_deliveries;
UPDATE schema_version SET version = 100;
-- +goose StatementEnd
