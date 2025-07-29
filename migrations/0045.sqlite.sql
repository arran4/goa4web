-- Allow direct emails in queue
ALTER TABLE pending_emails
    MODIFY COLUMN to_user_id INT NULL,
    ADD COLUMN direct_email TINYINT(1) NOT NULL DEFAULT 0;

-- Update schema version
UPDATE schema_version SET version = 45 WHERE version = 44;
