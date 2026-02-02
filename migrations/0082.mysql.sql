ALTER TABLE preferences ADD COLUMN weekly_digest_day INT DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN weekly_digest_hour INT DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN last_weekly_digest_sent_at DATETIME DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN monthly_digest_day INT DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN monthly_digest_hour INT DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN last_monthly_digest_sent_at DATETIME DEFAULT NULL;

CREATE TABLE scheduler_state (
    task_name VARCHAR(64) NOT NULL PRIMARY KEY,
    last_run_at DATETIME DEFAULT NULL,
    metadata TEXT DEFAULT NULL
);

-- Record upgrade to schema version 82
UPDATE schema_version SET version = 82 WHERE version = 81;
