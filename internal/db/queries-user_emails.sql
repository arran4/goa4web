-- name: InsertUserEmail :exec
INSERT INTO user_emails (user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListUserEmailsForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT ue.id, ue.user_id, ue.email, ue.verified_at, ue.last_verification_code, ue.verification_expires_at, ue.notification_priority
FROM user_emails ue
WHERE ue.user_id = sqlc.arg(user_id)
  AND (
      sqlc.arg(lister_id) = ue.user_id
      OR EXISTS (
          SELECT 1
          FROM role_ids ri
          JOIN roles r ON r.id = ri.id
          WHERE r.is_admin = 1
      )
  );

-- name: AdminListUserEmails :many
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE user_id = sqlc.arg(user_id)
ORDER BY notification_priority DESC, id;

-- name: SystemListVerifiedEmailsByUserID :many
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE user_id = sqlc.arg(user_id) AND verified_at IS NOT NULL
ORDER BY notification_priority DESC, id;

-- name: SystemListAllUserEmails :many
SELECT user_id, email, verified_at
FROM user_emails
ORDER BY user_id, email;

-- name: GetUserEmailByEmail :one
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE email = ?;

-- name: GetUserEmailByID :one
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE id = ?;

-- name: SystemMarkUserEmailVerified :exec
UPDATE user_emails
SET verified_at = ?, last_verification_code = NULL, verification_expires_at = NULL
WHERE id = ?;


-- name: SetNotificationPriorityForLister :exec
UPDATE user_emails SET notification_priority = sqlc.arg(notification_priority)
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(lister_id);

-- name: GetNotificationEmailByUserID :one
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE user_id = ? AND verified_at IS NOT NULL
ORDER BY notification_priority DESC, id
LIMIT 1;

-- name: DeleteUserEmailForOwner :exec
DELETE FROM user_emails WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(owner_id);

-- name: SetVerificationCodeForLister :exec
UPDATE user_emails
SET last_verification_code = sqlc.arg(last_verification_code),
    verification_expires_at = sqlc.arg(verification_expires_at)
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(lister_id);

-- name: GetUserEmailByCode :one
SELECT id, user_id, email, verified_at, last_verification_code, verification_expires_at, notification_priority
FROM user_emails
WHERE last_verification_code = ?;

-- name: GetMaxNotificationPriority :one
SELECT COALESCE(MAX(notification_priority),0) AS maxp FROM user_emails WHERE user_id = ?;

-- name: SystemDeleteUserEmailsByEmailExceptID :exec
DELETE FROM user_emails WHERE email = sqlc.arg(email) AND id != sqlc.arg(id);
