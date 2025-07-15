-- name: InsertPendingEmail :exec
INSERT INTO pending_emails (to_user_id, body)
VALUES (?, ?);

-- name: FetchPendingEmails :many
SELECT id, to_user_id, body, error_count
FROM pending_emails
WHERE sent_at IS NULL
ORDER BY id
LIMIT ?;

-- name: MarkEmailSent :exec
UPDATE pending_emails SET sent_at = NOW() WHERE id = ?;

-- name: ListUnsentPendingEmails :many
SELECT id, to_user_id, body, error_count, created_at
FROM pending_emails
WHERE sent_at IS NULL
ORDER BY id;

-- name: GetPendingEmailByID :one
SELECT id, to_user_id, body, error_count
FROM pending_emails
WHERE id = ?;

-- name: DeletePendingEmail :exec
DELETE FROM pending_emails WHERE id = ?;

-- name: IncrementEmailError :exec
UPDATE pending_emails SET error_count = error_count + 1 WHERE id = ?;

-- name: GetPendingEmailErrorCount :one
SELECT error_count FROM pending_emails WHERE id = ?;
