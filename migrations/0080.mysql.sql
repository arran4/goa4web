ALTER TABLE preferences ADD COLUMN daily_digest_hour INT DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN daily_digest_mark_read TINYINT(1) NOT NULL DEFAULT 0;
ALTER TABLE preferences ADD COLUMN last_digest_sent_at DATETIME DEFAULT NULL;

-- Record upgrade to schema version 80
UPDATE schema_version SET version = 80 WHERE version = 79;
