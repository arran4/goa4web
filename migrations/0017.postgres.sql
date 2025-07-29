-- Rename occured column to occurred on siteNews table
ALTER TABLE siteNews CHANGE COLUMN occured occurred datetime DEFAULT NULL;

-- Drop unused pending email columns and bump version
ALTER TABLE pending_emails
    DROP COLUMN IF EXISTS  subject,
    DROP COLUMN IF EXISTS html_body;

-- Record upgrade to schema version 17
UPDATE schema_version SET version = 17 WHERE version = 16;
