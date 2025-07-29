-- Add error_count column to email queue
ALTER TABLE pending_emails
    ADD COLUMN IF NOT EXISTS error_count INT NOT NULL DEFAULT 0;

-- Record upgrade to schema version 14
UPDATE schema_version SET version = 14 WHERE version = 13;
