-- System query only used internally
-- name: SystemInsertDeadLetter :exec
INSERT INTO dead_letters (message) VALUES (?);

-- name: SystemListDeadLetters :many
SELECT id, message, created_at FROM dead_letters
ORDER BY id DESC
LIMIT ?;

-- name: SystemDeleteDeadLetter :exec
DELETE FROM dead_letters WHERE id = ?;

-- name: SystemPurgeDeadLettersBefore :exec
DELETE FROM dead_letters WHERE created_at < ?;

-- name: SystemCountDeadLetters :one
SELECT COUNT(*) FROM dead_letters;

-- name: SystemLatestDeadLetter :one
SELECT MAX(created_at) FROM dead_letters;

-- name: SystemGetDeadLetter :one
SELECT id, message, created_at FROM dead_letters WHERE id = ?;

-- name: SystemUpdateDeadLetter :exec
UPDATE dead_letters SET message = ? WHERE id = ?;
