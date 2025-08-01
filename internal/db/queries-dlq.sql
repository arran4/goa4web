-- name: InsertDeadLetterSystem :exec
INSERT INTO dead_letters (message) VALUES (?);

-- name: ListDeadLettersForAdmin :many
SELECT id, message, created_at FROM dead_letters
ORDER BY id DESC
LIMIT ?;

-- name: DeleteDeadLetterForAdmin :exec
DELETE FROM dead_letters WHERE id = ?;

-- name: PurgeDeadLettersBeforeForAdmin :exec
DELETE FROM dead_letters WHERE created_at < ?;

-- name: CountDeadLettersSystem :one
SELECT COUNT(*) FROM dead_letters;

-- name: LatestDeadLetterForAdmin :one
SELECT MAX(created_at) FROM dead_letters;
