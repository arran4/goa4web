-- +goose Up
ALTER TABLE preferences ADD COLUMN weekly_digest_day INT DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN weekly_digest_hour INT DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN last_weekly_digest_sent_at DATETIME DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN monthly_digest_day INT DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN monthly_digest_hour INT DEFAULT NULL;
ALTER TABLE preferences ADD COLUMN last_monthly_digest_sent_at DATETIME DEFAULT NULL;

CREATE TABLE IF NOT EXISTS scheduler_state (
task_name TEXT NOT NULL PRIMARY KEY,
last_run_at DATETIME DEFAULT NULL,
metadata TEXT DEFAULT NULL
);
UPDATE schema_version SET version = 82;