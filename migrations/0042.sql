-- Allow non-unique unverified emails
ALTER TABLE user_emails
    DROP INDEX user_emails_email_idx,
    ADD UNIQUE KEY user_emails_email_code_idx (email(255), last_verification_code);

-- Update schema version
UPDATE schema_version SET version = 42 WHERE version = 41;
