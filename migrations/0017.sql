-- Drop unused pending email columns and bump version
ALTER TABLE pending_emails
    DROP COLUMN subject,
    DROP COLUMN html_body;

UPDATE schema_version SET version = 17 WHERE version = 16;
