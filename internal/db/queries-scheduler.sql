-- name: GetSchedulerState :one
SELECT task_name, last_run_at, metadata
FROM scheduler_state
WHERE task_name = sqlc.arg(task_name);

-- name: UpsertSchedulerState :exec
INSERT INTO scheduler_state (task_name, last_run_at, metadata)
VALUES (sqlc.arg(task_name), sqlc.arg(last_run_at), sqlc.arg(metadata))
ON DUPLICATE KEY UPDATE
    last_run_at = VALUES(last_run_at),
    metadata = VALUES(metadata);
