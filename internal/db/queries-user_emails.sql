-- name: InsertUserEmail :exec
INSERT INTO user_emails (user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetUserEmailsByUserID :many
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE user_id = ?;

-- name: ListVerifiedEmailsByUserID :many
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE user_id = ? AND verified_at IS NOT NULL
ORDER BY notification_priority DESC, id;

-- name: GetUserEmailByEmail :one
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE email = ?;

-- name: GetUserEmailByID :one
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE id = ?;

-- name: UpdateUserEmailVerification :exec
UPDATE user_emails
SET verified_at = ?, last_verification_code = NULL, verification_expires_at = NULL
WHERE id = ?;

-- name: ClearNotificationPriority :exec
UPDATE user_emails SET notification_priority = 0 WHERE user_id = ?;

-- name: SetNotificationPriority :exec
UPDATE user_emails SET notification_priority = ? WHERE id = ?;

-- name: GetNotificationEmailByUserID :one
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE user_id = ? AND verified_at IS NOT NULL
ORDER BY notification_priority DESC, id
LIMIT 1;

-- name: DeleteUserEmail :exec
DELETE FROM user_emails WHERE id = ?;

-- name: SetVerificationCode :exec
UPDATE user_emails SET last_verification_code = ?, verification_expires_at = ? WHERE id = ?;

-- name: GetUserEmailByCode :one
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE last_verification_code = ?;

-- name: GetMaxNotificationPriority :one
SELECT COALESCE(MAX(notification_priority),0) AS maxp FROM user_emails WHERE user_id = ?;

-- name: DeleteUserEmailsByEmailExceptID :exec
DELETE FROM user_emails WHERE email = ? AND id != ?;

