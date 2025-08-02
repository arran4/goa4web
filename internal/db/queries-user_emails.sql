-- name: InsertUserEmail :exec
INSERT INTO user_emails (user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetUserEmailsByUserID :many
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT ue.id, ue.user_id, ue.email, ue.verified_at, ue.last_verification_code, ue.verification_expires_at, ue.notification_priority
FROM user_emails ue
WHERE ue.user_id = sqlc.arg(user_id)
  AND (
      sqlc.arg(viewer_id) = ue.user_id
      OR EXISTS (
          SELECT 1
          FROM role_ids ri
          JOIN roles r ON r.id = ri.id
          WHERE r.is_admin = 1
      )
  );

-- name: GetUserEmailsByUserIDAdmin :many
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE user_id = ?
ORDER BY notification_priority DESC, id;

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

