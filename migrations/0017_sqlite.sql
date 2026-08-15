-- +goose Up
-- ALTER TABLE siteNews CHANGE COLUMN occured occurred datetime DEFAULT NULL;

-- ALTER TABLE pending_emails
    DROP COLUMN IF EXISTS  subject,
    DROP COLUMN IF EXISTS html_body;