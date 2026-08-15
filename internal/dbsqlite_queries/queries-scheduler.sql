-- name: GetSchedulerState :one
SELECT task_name, last_run_at, metadata
FROM scheduler_state
WHERE task_name = sqlc.arg(task_name);

-- name: UpsertSchedulerState :exec
INSERT INTO scheduler_state (task_name, last_run_at, metadata)
VALUES (sqlc.arg(task_name), sqlc.arg(last_run_at), sqlc.arg(metadata))
ON CONFLICT(task_name) DO UPDATE SET last_run_at = excluded.last_run_at, metadata = excluded.metadata;
