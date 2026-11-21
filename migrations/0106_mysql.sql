-- +goose Up
ALTER TABLE preferences ADD COLUMN optional_email_generation BIGINT NOT NULL DEFAULT 0;
CREATE TABLE forum_reply_email_activities (
    recipient_id INT NOT NULL,
    activity_position BIGINT NOT NULL,
    thread_id INT NOT NULL,
    window_end BIGINT NOT NULL,
    preference_generation BIGINT NOT NULL DEFAULT 0,
    claim_count INT NOT NULL DEFAULT 0,
    lease_token VARCHAR(36) DEFAULT NULL,
    lease_until BIGINT NOT NULL DEFAULT 0,
    attempted_at BIGINT DEFAULT NULL,
    sent_at BIGINT DEFAULT NULL,
    PRIMARY KEY (recipient_id, activity_position)
);
CREATE INDEX forum_reply_email_due ON forum_reply_email_activities (attempted_at, window_end, lease_until, recipient_id);
CREATE INDEX forum_reply_email_window ON forum_reply_email_activities (recipient_id, window_end, lease_token);
UPDATE schema_version SET version = 106;

-- +goose Down
-- Dropping the queue loses pending optional reply activity and its replay history.
DROP TABLE forum_reply_email_activities;
ALTER TABLE preferences DROP COLUMN optional_email_generation;
UPDATE schema_version SET version = 105;
