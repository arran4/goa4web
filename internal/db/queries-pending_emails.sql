-- name: InsertPendingEmail :exec
INSERT INTO pending_emails (to_user_id, body, direct_email)
VALUES (?, ?, ?);

-- name: SystemListPendingEmails :many
SELECT id, to_user_id, body, error_count, direct_email
FROM pending_emails
WHERE sent_at IS NULL
ORDER BY id
LIMIT ? OFFSET ?;

-- name: SystemMarkPendingEmailSent :exec
UPDATE pending_emails SET sent_at = NOW() WHERE id = ?;

-- name: AdminListUnsentPendingEmails :many
-- admin task
SELECT pe.id, pe.to_user_id, pe.body, pe.error_count, pe.created_at, pe.direct_email
FROM pending_emails pe
LEFT JOIN preferences p ON pe.to_user_id = p.users_idusers
LEFT JOIN user_roles ur ON pe.to_user_id = ur.users_idusers
LEFT JOIN roles r ON ur.role_id = r.id
WHERE pe.sent_at IS NULL
  AND (sqlc.narg(language_id) IS NULL OR p.language_idlanguage = sqlc.narg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name))
ORDER BY pe.id;

-- name: AdminGetPendingEmailByID :one
-- admin task
SELECT id, to_user_id, body, error_count, direct_email
FROM pending_emails
WHERE id = ?;

-- name: AdminDeletePendingEmail :exec
-- admin task
DELETE FROM pending_emails WHERE id = ?;

-- name: SystemIncrementPendingEmailError :exec
UPDATE pending_emails SET error_count = error_count + 1 WHERE id = ?;

-- name: GetPendingEmailErrorCount :one
SELECT error_count FROM pending_emails WHERE id = ?;

-- name: AdminListSentEmails :many
-- admin task
SELECT pe.id, pe.to_user_id, pe.body, pe.error_count, pe.created_at, pe.sent_at, pe.direct_email
FROM pending_emails pe
LEFT JOIN preferences p ON pe.to_user_id = p.users_idusers
LEFT JOIN user_roles ur ON pe.to_user_id = ur.users_idusers
LEFT JOIN roles r ON ur.role_id = r.id
WHERE pe.sent_at IS NOT NULL
  AND (sqlc.narg(language_id) IS NULL OR p.language_idlanguage = sqlc.narg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name))
ORDER BY pe.sent_at DESC
LIMIT ? OFFSET ?;

-- name: AdminListFailedEmails :many
-- admin task
SELECT pe.id, pe.to_user_id, pe.body, pe.error_count, pe.created_at, pe.direct_email
FROM pending_emails pe
LEFT JOIN preferences p ON pe.to_user_id = p.users_idusers
LEFT JOIN user_roles ur ON pe.to_user_id = ur.users_idusers
LEFT JOIN roles r ON ur.role_id = r.id
WHERE pe.sent_at IS NULL AND pe.error_count > 0
  AND (sqlc.narg(language_id) IS NULL OR p.language_idlanguage = sqlc.narg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name))
ORDER BY pe.id
LIMIT ? OFFSET ?;
