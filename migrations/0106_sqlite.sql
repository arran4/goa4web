-- +goose Up
ALTER TABLE preferences ADD COLUMN optional_email_generation INTEGER NOT NULL DEFAULT 0;
CREATE TABLE forum_reply_email_activities (
    recipient_id INTEGER NOT NULL,
    activity_position INTEGER NOT NULL,
    thread_id INTEGER NOT NULL,
    window_end INTEGER NOT NULL,
    preference_generation INTEGER NOT NULL DEFAULT 0,
    claim_count INTEGER NOT NULL DEFAULT 0,
    lease_token TEXT DEFAULT NULL,
    lease_until INTEGER NOT NULL DEFAULT 0,
    attempted_at INTEGER DEFAULT NULL,
    sent_at INTEGER DEFAULT NULL,
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
