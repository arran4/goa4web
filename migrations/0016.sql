-- Replace to_email with to_user_id referencing users
ALTER TABLE pending_emails
    DROP COLUMN to_email,
    ADD COLUMN to_user_id INT NOT NULL DEFAULT 0;

-- Update schema version
UPDATE schema_version SET version = 16 WHERE version = 15;
