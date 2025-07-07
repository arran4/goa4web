-- name: InsertDeadLetter :exec
INSERT INTO dead_letters (message) VALUES (?);

-- name: ListDeadLetters :many
SELECT id, message, created_at FROM dead_letters
ORDER BY id DESC
LIMIT ?;

-- name: DeleteDeadLetter :exec
DELETE FROM dead_letters WHERE id = ?;

-- name: PurgeDeadLettersBefore :exec
DELETE FROM dead_letters WHERE created_at < ?;

-- name: CountDeadLetters :one
SELECT COUNT(*) FROM dead_letters;
