-- Replace to_email with to_user_id referencing users
ALTER TABLE pending_emails
    DROP COLUMN to_email,
    ADD COLUMN to_user_id INT NOT NULL DEFAULT 0;

-- Update schema version
-- Remove obsolete columns from subscriptions table
-- Remove obsolete item_type and target_id columns from subscriptions
ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS item_type,
    DROP COLUMN IF EXISTS target_id;

-- Remove html_body column from pending_emails table
ALTER TABLE pending_emails DROP COLUMN IF EXISTS html_body;

-- Record upgrade to schema version 16
UPDATE schema_version SET version = 16 WHERE version = 15;
