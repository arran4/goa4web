-- +goose Up
-- +goose StatementBegin
CREATE TABLE durable_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_type TEXT NOT NULL,
    dedupe_key TEXT NOT NULL,
    active_dedupe_key TEXT DEFAULT NULL,
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL DEFAULT 0,
    target_key TEXT NOT NULL DEFAULT '',
    target_revision INTEGER NOT NULL DEFAULT 0,
    needs_rerun INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 5,
    run_after DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    lease_owner TEXT DEFAULT NULL,
    lease_token TEXT DEFAULT NULL,
    lease_expires_at DATETIME DEFAULT NULL,
    last_error TEXT DEFAULT NULL,
    created_by_user_id INTEGER DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME DEFAULT NULL
);
CREATE UNIQUE INDEX uq_durable_jobs_active_dedupe ON durable_jobs (active_dedupe_key);
CREATE INDEX idx_durable_jobs_claim ON durable_jobs (status, run_after, lease_expires_at);
CREATE INDEX idx_durable_jobs_dedupe ON durable_jobs (dedupe_key);
CREATE INDEX idx_durable_jobs_completed ON durable_jobs (status, completed_at);

CREATE TABLE durable_job_cooldowns (
    dedupe_key TEXT NOT NULL PRIMARY KEY,
    cooldown_until DATETIME NOT NULL,
    last_error TEXT DEFAULT NULL,
    last_notified_at DATETIME DEFAULT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_durable_job_cooldowns_until ON durable_job_cooldowns (cooldown_until);

CREATE TABLE search_dirty_targets (
    target_type TEXT NOT NULL,
    target_id INTEGER NOT NULL,
    revision INTEGER NOT NULL DEFAULT 1,
    indexed_revision INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (target_type, target_id)
);

UPDATE schema_version SET version = 104;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE search_dirty_targets;
DROP TABLE durable_job_cooldowns;
DROP TABLE durable_jobs;
UPDATE schema_version SET version = 103;
-- +goose StatementEnd
