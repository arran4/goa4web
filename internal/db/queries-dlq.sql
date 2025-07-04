-- name: InsertWorkerError :exec
INSERT INTO worker_errors (message) VALUES (?);

-- name: ListWorkerErrors :many
SELECT id, message, created_at FROM worker_errors
ORDER BY id DESC
LIMIT ?;

-- name: DeleteWorkerError :exec
DELETE FROM worker_errors WHERE id = ?;

-- name: PurgeWorkerErrorsBefore :exec
DELETE FROM worker_errors WHERE created_at < ?;

-- name: CountWorkerErrors :one
SELECT COUNT(*) FROM worker_errors;
