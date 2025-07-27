-- name: InsertPendingEmail :exec
INSERT INTO pending_emails (to_user_id, body, direct_email)
VALUES (?, ?, ?);

-- name: FetchPendingEmails :many
SELECT id, to_user_id, body, error_count, direct_email
FROM pending_emails
WHERE sent_at IS NULL
ORDER BY id
LIMIT ?;

-- name: MarkEmailSent :exec
UPDATE pending_emails SET sent_at = NOW() WHERE id = ?;

-- name: ListUnsentPendingEmails :many
SELECT id, to_user_id, body, error_count, created_at, direct_email
FROM pending_emails
WHERE sent_at IS NULL
ORDER BY id;

-- name: GetPendingEmailByID :one
SELECT id, to_user_id, body, error_count, direct_email
FROM pending_emails
WHERE id = ?;

-- name: DeletePendingEmail :exec
DELETE FROM pending_emails WHERE id = ?;

-- name: IncrementEmailError :exec
UPDATE pending_emails SET error_count = error_count + 1 WHERE id = ?;

-- name: GetPendingEmailErrorCount :one
SELECT error_count FROM pending_emails WHERE id = ?;

-- name: ListSentEmails :many
SELECT id, to_user_id, body, error_count, created_at, sent_at, direct_email
FROM pending_emails
WHERE sent_at IS NOT NULL
ORDER BY sent_at DESC
LIMIT ? OFFSET ?;

-- name: ListFailedEmails :many
SELECT id, to_user_id, body, error_count, created_at, direct_email
FROM pending_emails
WHERE sent_at IS NULL AND error_count > 0
ORDER BY id
LIMIT ? OFFSET ?;
