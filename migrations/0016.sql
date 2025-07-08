-- Remove html_body column from pending_emails table
ALTER TABLE pending_emails DROP COLUMN IF EXISTS html_body;

-- Record upgrade to schema version 16
UPDATE schema_version SET version = 16 WHERE version = 15;
