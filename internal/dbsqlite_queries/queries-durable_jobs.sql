-- name: SystemInsertDurableJob :execresult
INSERT INTO durable_jobs (
    job_type,
    dedupe_key,
    active_dedupe_key,
    target_type,
    target_id,
    target_key,
    target_revision,
    needs_rerun,
    status,
    attempts,
    max_attempts,
    run_after,
    created_by_user_id,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, 0, 'pending', 0, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
);

-- name: SystemBumpActiveDurableJob :execrows
UPDATE durable_jobs
SET
    target_revision = CASE WHEN target_revision < ? THEN ? ELSE target_revision END,
    needs_rerun = CASE WHEN status = 'running' THEN 1 ELSE needs_rerun END,
    run_after = CASE WHEN status = 'pending' AND run_after > ? THEN ? ELSE run_after END,
    updated_at = CURRENT_TIMESTAMP
WHERE active_dedupe_key = ?
  AND status IN ('pending', 'running');

-- name: SystemGetNextClaimableDurableJob :one
SELECT
    id, job_type, dedupe_key, active_dedupe_key, target_type, target_id,
    target_key, target_revision, needs_rerun, status, attempts, max_attempts,
    run_after, lease_owner, lease_token, lease_expires_at, last_error,
    created_by_user_id, created_at, updated_at, completed_at
FROM durable_jobs
WHERE (status = 'pending' AND run_after <= ?)
   OR (status = 'running' AND lease_expires_at IS NOT NULL AND lease_expires_at <= ?)
ORDER BY run_after ASC, id ASC
LIMIT 1;

-- name: SystemGetNextClaimableDurableJobByTypes :one
SELECT
    id, job_type, dedupe_key, active_dedupe_key, target_type, target_id,
    target_key, target_revision, needs_rerun, status, attempts, max_attempts,
    run_after, lease_owner, lease_token, lease_expires_at, last_error,
    created_by_user_id, created_at, updated_at, completed_at
FROM durable_jobs
WHERE job_type IN (sqlc.slice('job_types'))
  AND (
      (status = 'pending' AND run_after <= ?)
      OR (status = 'running' AND lease_expires_at IS NOT NULL AND lease_expires_at <= ?)
  )
ORDER BY run_after ASC, id ASC
LIMIT 1;

-- name: SystemClaimDurableJobByID :execrows
UPDATE durable_jobs
SET
    status = 'running',
    attempts = attempts + 1,
    lease_owner = ?,
    lease_token = ?,
    lease_expires_at = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND (
      (status = 'pending' AND run_after <= ?)
      OR (status = 'running' AND lease_expires_at IS NOT NULL AND lease_expires_at <= ?)
  );

-- name: SystemRenewDurableJobLease :execrows
UPDATE durable_jobs
SET
    lease_expires_at = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND status = 'running'
  AND lease_token = ?;

-- name: SystemCompleteDurableJobFenced :execrows
UPDATE durable_jobs
SET
    status = CASE WHEN needs_rerun = 1 THEN 'pending' ELSE 'completed' END,
    active_dedupe_key = CASE WHEN needs_rerun = 1 THEN dedupe_key ELSE NULL END,
    needs_rerun = 0,
    run_after = CASE WHEN needs_rerun = 1 THEN ? ELSE run_after END,
    lease_owner = NULL,
    lease_token = NULL,
    lease_expires_at = NULL,
    last_error = NULL,
    completed_at = CASE WHEN needs_rerun = 1 THEN NULL ELSE ? END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND status = 'running'
  AND lease_token = ?;

-- name: SystemRetryDurableJobFenced :execrows
UPDATE durable_jobs
SET
    status = 'pending',
    run_after = ?,
    lease_owner = NULL,
    lease_token = NULL,
    lease_expires_at = NULL,
    last_error = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND status = 'running'
  AND lease_token = ?;

-- name: SystemFailDurableJobFenced :execrows
UPDATE durable_jobs
SET
    status = 'failed',
    active_dedupe_key = NULL,
    needs_rerun = 0,
    lease_owner = NULL,
    lease_token = NULL,
    lease_expires_at = NULL,
    last_error = ?,
    completed_at = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND status = 'running'
  AND lease_token = ?;

-- name: SystemReleaseDurableJobOnShutdownFenced :execrows
UPDATE durable_jobs
SET
    status = 'pending',
    attempts = CASE WHEN attempts > 0 THEN attempts - 1 ELSE 0 END,
    run_after = ?,
    lease_owner = NULL,
    lease_token = NULL,
    lease_expires_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND status = 'running'
  AND lease_token = ?;

-- name: SystemGetDurableJobByID :one
SELECT
    id, job_type, dedupe_key, active_dedupe_key, target_type, target_id,
    target_key, target_revision, needs_rerun, status, attempts, max_attempts,
    run_after, lease_owner, lease_token, lease_expires_at, last_error,
    created_by_user_id, created_at, updated_at, completed_at
FROM durable_jobs
WHERE id = ?
LIMIT 1;

-- name: AdminListDurableJobs :many
SELECT
    id, job_type, dedupe_key, active_dedupe_key, target_type, target_id,
    target_key, target_revision, needs_rerun, status, attempts, max_attempts,
    run_after, lease_owner, lease_token, lease_expires_at, last_error,
    created_by_user_id, created_at, updated_at, completed_at
FROM durable_jobs
WHERE (? = '' OR status = ?)
  AND (? = '' OR job_type = ?)
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: AdminRetryDurableJob :execrows
UPDATE durable_jobs
SET
    status = 'pending',
    active_dedupe_key = dedupe_key,
    attempts = 0,
    run_after = CURRENT_TIMESTAMP,
    lease_owner = NULL,
    lease_token = NULL,
    lease_expires_at = NULL,
    last_error = NULL,
    completed_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND status IN ('failed', 'cancelled', 'completed');

-- name: AdminCancelDurableJob :execrows
UPDATE durable_jobs
SET
    status = 'cancelled',
    active_dedupe_key = NULL,
    needs_rerun = 0,
    lease_owner = NULL,
    lease_token = NULL,
    lease_expires_at = NULL,
    completed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
  AND status IN ('pending', 'running');

-- name: AdminPurgeTerminalDurableJobs :execrows
DELETE FROM durable_jobs
WHERE status IN ('completed', 'failed', 'cancelled')
  AND (? = '' OR status = ?);

-- name: SystemPruneTerminalDurableJobs :execrows
DELETE FROM durable_jobs
WHERE durable_jobs.status IN ('completed', 'failed', 'cancelled')
  AND (
    durable_jobs.completed_at < sqlc.arg(cutoff)
    OR durable_jobs.id IN (
        SELECT ranked.id FROM (
            SELECT history.id, ROW_NUMBER() OVER (PARTITION BY history.job_type ORDER BY history.completed_at DESC, history.id DESC) AS history_rank
            FROM durable_jobs history
            WHERE history.status IN ('completed', 'failed', 'cancelled')
        ) ranked
        WHERE history_rank > CAST(sqlc.arg(max_per_kind) AS INTEGER)
    )
  );

-- name: SystemGetJobCooldown :one
SELECT dedupe_key, cooldown_until, last_error, last_notified_at, updated_at
FROM durable_job_cooldowns
WHERE dedupe_key = ?
LIMIT 1;

-- name: SystemUpsertJobCooldown :exec
INSERT INTO durable_job_cooldowns (dedupe_key, cooldown_until, last_error, last_notified_at, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(dedupe_key) DO UPDATE SET
    cooldown_until = excluded.cooldown_until,
    last_error = excluded.last_error,
    last_notified_at = COALESCE(excluded.last_notified_at, durable_job_cooldowns.last_notified_at),
    updated_at = CURRENT_TIMESTAMP;

-- name: SystemClearJobCooldownWindow :exec
UPDATE durable_job_cooldowns
SET cooldown_until = ?, updated_at = CURRENT_TIMESTAMP
WHERE dedupe_key = ?;

-- name: SystemDeleteJobCooldown :exec
DELETE FROM durable_job_cooldowns
WHERE dedupe_key = ?;

-- name: SystemBumpSearchDirtyTarget :exec
INSERT INTO search_dirty_targets (target_type, target_id, revision, indexed_revision, updated_at)
VALUES (?, ?, 1, 0, CURRENT_TIMESTAMP)
ON CONFLICT(target_type, target_id) DO UPDATE SET
    revision = search_dirty_targets.revision + 1,
    updated_at = CURRENT_TIMESTAMP;

-- name: SystemGetSearchDirtyTarget :one
SELECT target_type, target_id, revision, indexed_revision, updated_at
FROM search_dirty_targets
WHERE target_type = ? AND target_id = ?
LIMIT 1;

-- name: SystemMarkSearchTargetIndexed :execrows
UPDATE search_dirty_targets
SET indexed_revision = ?, updated_at = CURRENT_TIMESTAMP
WHERE target_type = ?
  AND target_id = ?
  AND indexed_revision < ?;

-- name: SystemListActiveAdminUserIDs :many
SELECT DISTINCT u.idusers
FROM users u
JOIN user_roles ur ON ur.users_idusers = u.idusers
JOIN roles r ON ur.role_id = r.id
WHERE r.is_admin = 1
  AND u.deleted_at IS NULL
ORDER BY u.idusers;

-- name: SystemGetCommentTextForSearchIndex :one
SELECT idcomments, text
FROM comments
WHERE idcomments = ? AND deleted_at IS NULL
LIMIT 1;

-- name: SystemGetWritingTextForSearchIndex :one
SELECT idwriting, title, writing, abstract
FROM writing
WHERE idwriting = ? AND deleted_at IS NULL
LIMIT 1;

-- name: SystemGetLinkerTextForSearchIndex :one
SELECT id, title, url, description
FROM linker
WHERE id = ? AND deleted_at IS NULL
LIMIT 1;

-- name: SystemGetImagePostTextForSearchIndex :one
SELECT idimagepost, description
FROM imagepost
WHERE idimagepost = ? AND deleted_at IS NULL
LIMIT 1;
