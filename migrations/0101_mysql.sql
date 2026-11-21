-- +goose Up
-- +goose StatementBegin
CREATE TABLE digest_deliveries (
    id INT NOT NULL AUTO_INCREMENT,
    user_id INT NOT NULL,
    dismiss_notifications TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at DATETIME DEFAULT NULL,
    PRIMARY KEY (id),
    KEY digest_deliveries_user_idx (user_id)
);

CREATE TABLE digest_delivery_notifications (
    digest_delivery_id INT NOT NULL,
    notification_id INT NOT NULL,
    PRIMARY KEY (digest_delivery_id, notification_id),
    KEY digest_delivery_notifications_notification_idx (notification_id)
);

ALTER TABLE pending_emails
    ADD COLUMN digest_delivery_id INT NULL,
    ADD UNIQUE KEY pending_emails_digest_delivery_idx (digest_delivery_id);

UPDATE schema_version SET version = 101;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pending_emails
    DROP INDEX pending_emails_digest_delivery_idx,
    DROP COLUMN digest_delivery_id;
DROP TABLE digest_delivery_notifications;
DROP TABLE digest_deliveries;
UPDATE schema_version SET version = 100;
-- +goose StatementEnd
