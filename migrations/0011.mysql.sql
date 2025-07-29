-- Add method and pattern columns for subscription types and URI patterns
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS method VARCHAR(16) NOT NULL DEFAULT 'internal',
    ADD COLUMN IF NOT EXISTS pattern VARCHAR(255) NOT NULL DEFAULT '';

-- Create worker_errors table for async error records
CREATE TABLE IF NOT EXISTS worker_errors (
    id INT NOT NULL AUTO_INCREMENT,
    message TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

-- Add html_body column for multipart emails
ALTER TABLE pending_emails ADD COLUMN IF NOT EXISTS html_body text;

-- Record upgrade to schema version 11
UPDATE schema_version SET version = 11 WHERE version = 10;
