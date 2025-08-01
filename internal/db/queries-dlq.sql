-- name: SystemInsertDeadLetter :exec
INSERT INTO dead_letters (message) VALUES (?);

-- name: AdminListDeadLetters :many
SELECT id, message, created_at FROM dead_letters
ORDER BY id DESC
LIMIT ?;

-- name: AdminDeleteDeadLetter :exec
DELETE FROM dead_letters WHERE id = ?;

-- name: AdminPurgeDeadLettersBefore :exec
DELETE FROM dead_letters WHERE created_at < ?;

-- name: SystemCountDeadLetters :one
SELECT COUNT(*) FROM dead_letters;

-- name: AdminLatestDeadLetter :one
SELECT MAX(created_at) FROM dead_letters;
