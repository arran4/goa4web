-- +goose Up
-- ALTER TABLE pending_emails
    DROP COLUMN to_email,
    ADD COLUMN to_user_id INT NOT NULL DEFAULT 0;

-- ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS item_type,
    DROP COLUMN IF EXISTS target_id;

-- ALTER TABLE pending_emails DROP COLUMN IF EXISTS html_body;