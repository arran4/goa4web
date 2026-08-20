-- name: InsertPendingEmail :exec
INSERT INTO pending_emails (to_user_id, body, direct_email)
VALUES (?, ?, ?);

-- name: SystemListPendingEmails :many
SELECT id, to_user_id, body, error_count, direct_email, created_at
FROM pending_emails
WHERE sent_at IS NULL
ORDER BY id
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: SystemMarkPendingEmailSent :exec
UPDATE pending_emails SET sent_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: AdminListUnsentPendingEmails :many
-- admin task
SELECT pe.id, pe.to_user_id, pe.body, pe.error_count, pe.created_at, pe.direct_email
FROM pending_emails pe
LEFT JOIN preferences p ON pe.to_user_id = p.users_idusers
LEFT JOIN user_roles ur ON pe.to_user_id = ur.users_idusers
LEFT JOIN roles r ON ur.role_id = r.id
WHERE pe.sent_at IS NULL
  AND (sqlc.arg(status) IS NULL
    OR (sqlc.arg(status) = 'pending' AND pe.error_count = 0)
    OR (sqlc.arg(status) = 'failed' AND pe.error_count > 0))
  AND (sqlc.arg(provider) IS NULL
    OR (sqlc.arg(provider) = 'direct' AND pe.direct_email = 1)
    OR (sqlc.arg(provider) = 'user' AND pe.direct_email = 0 AND pe.to_user_id IS NOT NULL AND pe.to_user_id <> 0)
    OR (sqlc.arg(provider) = 'userless' AND pe.direct_email = 0 AND (pe.to_user_id IS NULL OR pe.to_user_id = 0)))
  AND (sqlc.arg(created_before) IS NULL OR pe.created_at <= sqlc.arg(created_before))
  AND (sqlc.arg(language_id) IS NULL OR p.language_id = sqlc.arg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name))
ORDER BY pe.id
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminCountUnsentPendingEmails :one
-- admin task
SELECT COUNT(*)
FROM pending_emails pe
LEFT JOIN preferences p ON pe.to_user_id = p.users_idusers
LEFT JOIN user_roles ur ON pe.to_user_id = ur.users_idusers
LEFT JOIN roles r ON ur.role_id = r.id
WHERE pe.sent_at IS NULL
  AND (sqlc.arg(status) IS NULL
    OR (sqlc.arg(status) = 'pending' AND pe.error_count = 0)
    OR (sqlc.arg(status) = 'failed' AND pe.error_count > 0))
  AND (sqlc.arg(provider) IS NULL
    OR (sqlc.arg(provider) = 'direct' AND pe.direct_email = 1)
    OR (sqlc.arg(provider) = 'user' AND pe.direct_email = 0 AND pe.to_user_id IS NOT NULL AND pe.to_user_id <> 0)
    OR (sqlc.arg(provider) = 'userless' AND pe.direct_email = 0 AND (pe.to_user_id IS NULL OR pe.to_user_id = 0)))
  AND (sqlc.arg(created_before) IS NULL OR pe.created_at <= sqlc.arg(created_before))
  AND (sqlc.arg(language_id) IS NULL OR p.language_id = sqlc.arg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name));

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
  AND (sqlc.arg(provider) IS NULL
    OR (sqlc.arg(provider) = 'direct' AND pe.direct_email = 1)
    OR (sqlc.arg(provider) = 'user' AND pe.direct_email = 0 AND pe.to_user_id IS NOT NULL AND pe.to_user_id <> 0)
    OR (sqlc.arg(provider) = 'userless' AND pe.direct_email = 0 AND (pe.to_user_id IS NULL OR pe.to_user_id = 0)))
  AND (sqlc.arg(created_before) IS NULL OR pe.sent_at <= sqlc.arg(created_before))
  AND (sqlc.arg(language_id) IS NULL OR p.language_id = sqlc.arg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name))
ORDER BY pe.sent_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminCountSentEmails :one
-- admin task
SELECT COUNT(*)
FROM pending_emails pe
LEFT JOIN preferences p ON pe.to_user_id = p.users_idusers
LEFT JOIN user_roles ur ON pe.to_user_id = ur.users_idusers
LEFT JOIN roles r ON ur.role_id = r.id
WHERE pe.sent_at IS NOT NULL
  AND (sqlc.arg(provider) IS NULL
    OR (sqlc.arg(provider) = 'direct' AND pe.direct_email = 1)
    OR (sqlc.arg(provider) = 'user' AND pe.direct_email = 0 AND pe.to_user_id IS NOT NULL AND pe.to_user_id <> 0)
    OR (sqlc.arg(provider) = 'userless' AND pe.direct_email = 0 AND (pe.to_user_id IS NULL OR pe.to_user_id = 0)))
  AND (sqlc.arg(created_before) IS NULL OR pe.sent_at <= sqlc.arg(created_before))
  AND (sqlc.arg(language_id) IS NULL OR p.language_id = sqlc.arg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name));

-- name: AdminListSentEmailIDs :many
-- admin task
SELECT pe.id
FROM pending_emails pe
LEFT JOIN preferences p ON pe.to_user_id = p.users_idusers
LEFT JOIN user_roles ur ON pe.to_user_id = ur.users_idusers
LEFT JOIN roles r ON ur.role_id = r.id
WHERE pe.sent_at IS NOT NULL
  AND (sqlc.arg(provider) IS NULL
    OR (sqlc.arg(provider) = 'direct' AND pe.direct_email = 1)
    OR (sqlc.arg(provider) = 'user' AND pe.direct_email = 0 AND pe.to_user_id IS NOT NULL AND pe.to_user_id <> 0)
    OR (sqlc.arg(provider) = 'userless' AND pe.direct_email = 0 AND (pe.to_user_id IS NULL OR pe.to_user_id = 0)))
  AND (sqlc.arg(created_before) IS NULL OR pe.sent_at <= sqlc.arg(created_before))
  AND (sqlc.arg(language_id) IS NULL OR p.language_id = sqlc.arg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name))
ORDER BY pe.sent_at DESC;

-- name: AdminListFailedEmails :many
-- admin task
SELECT pe.id, pe.to_user_id, pe.body, pe.error_count, pe.created_at, pe.direct_email
FROM pending_emails pe
LEFT JOIN preferences p ON pe.to_user_id = p.users_idusers
LEFT JOIN user_roles ur ON pe.to_user_id = ur.users_idusers
LEFT JOIN roles r ON ur.role_id = r.id
WHERE pe.sent_at IS NULL AND pe.error_count > 0
  AND (sqlc.arg(language_id) IS NULL OR p.language_id = sqlc.arg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name))
ORDER BY pe.id
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminListFailedEmailIDs :many
-- admin task
SELECT pe.id
FROM pending_emails pe
LEFT JOIN preferences p ON pe.to_user_id = p.users_idusers
LEFT JOIN user_roles ur ON pe.to_user_id = ur.users_idusers
LEFT JOIN roles r ON ur.role_id = r.id
WHERE pe.sent_at IS NULL AND pe.error_count > 0
  AND (sqlc.arg(provider) IS NULL
    OR (sqlc.arg(provider) = 'direct' AND pe.direct_email = 1)
    OR (sqlc.arg(provider) = 'user' AND pe.direct_email = 0 AND pe.to_user_id IS NOT NULL AND pe.to_user_id <> 0)
    OR (sqlc.arg(provider) = 'userless' AND pe.direct_email = 0 AND (pe.to_user_id IS NULL OR pe.to_user_id = 0)))
  AND (sqlc.arg(created_before) IS NULL OR pe.created_at <= sqlc.arg(created_before))
  AND (sqlc.arg(language_id) IS NULL OR p.language_id = sqlc.arg(language_id))
  AND (sqlc.arg(role_name) IS NULL OR r.name = sqlc.arg(role_name))
ORDER BY pe.id;
