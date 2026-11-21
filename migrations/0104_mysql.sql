-- +goose Up
-- +goose StatementBegin
CREATE TABLE durable_jobs (
    id BIGINT NOT NULL AUTO_INCREMENT,
    job_type VARCHAR(64) NOT NULL,
    dedupe_key VARCHAR(255) NOT NULL,
    active_dedupe_key VARCHAR(255) DEFAULT NULL,
    target_type VARCHAR(64) NOT NULL,
    target_id BIGINT NOT NULL DEFAULT 0,
    target_key VARCHAR(512) NOT NULL DEFAULT '',
    target_revision BIGINT NOT NULL DEFAULT 0,
    needs_rerun TINYINT(1) NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    run_after DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    lease_owner VARCHAR(128) DEFAULT NULL,
    lease_token VARCHAR(64) DEFAULT NULL,
    lease_expires_at DATETIME DEFAULT NULL,
    last_error TEXT DEFAULT NULL,
    created_by_user_id INT DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    completed_at DATETIME DEFAULT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_durable_jobs_active_dedupe (active_dedupe_key),
    KEY idx_durable_jobs_claim (status, run_after, lease_expires_at),
    KEY idx_durable_jobs_dedupe (dedupe_key),
    KEY idx_durable_jobs_completed (status, completed_at)
);

CREATE TABLE durable_job_cooldowns (
    dedupe_key VARCHAR(255) NOT NULL,
    cooldown_until DATETIME NOT NULL,
    last_error TEXT DEFAULT NULL,
    last_notified_at DATETIME DEFAULT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (dedupe_key),
    KEY idx_durable_job_cooldowns_until (cooldown_until)
);

CREATE TABLE search_dirty_targets (
    target_type VARCHAR(64) NOT NULL,
    target_id INT NOT NULL,
    revision BIGINT NOT NULL DEFAULT 1,
    indexed_revision BIGINT NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
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
