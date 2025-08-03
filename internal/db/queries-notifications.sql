-- SystemInsertNotification stores an internal notification for a user.
-- Parameters:
--   UserID
--   Link
--   Message
-- name: SystemInsertNotification :exec
INSERT INTO notifications (users_idusers, link, message)
VALUES (sqlc.arg(user_id), sqlc.arg(link), sqlc.arg(message));

-- name: CountUnreadNotificationsForLister :one
SELECT COUNT(*) FROM notifications
WHERE users_idusers = sqlc.arg(lister_id) AND read_at IS NULL;

-- name: ListNotificationsForLister :many
SELECT id, users_idusers, link, message, created_at, read_at
FROM notifications
WHERE users_idusers = sqlc.arg(lister_id)
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListUnreadNotificationsForLister :many
SELECT id, users_idusers, link, message, created_at, read_at
FROM notifications
WHERE users_idusers = sqlc.arg(lister_id) AND read_at IS NULL
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: MarkNotificationReadForReader :exec
UPDATE notifications SET read_at = NOW()
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(reader_id);

-- name: AdminMarkNotificationRead :exec
UPDATE notifications SET read_at = NOW() WHERE id = sqlc.arg(ID);

-- name: MarkNotificationUnreadForReader :exec
UPDATE notifications SET read_at = NULL
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(reader_id);

-- name: AdminMarkNotificationUnread :exec
UPDATE notifications SET read_at = NULL WHERE id = sqlc.arg(ID);

-- name: AdminDeleteNotification :exec
DELETE FROM notifications WHERE id = sqlc.arg(id);

-- name: AdminGetNotification :one
SELECT id, users_idusers, link, message, created_at, read_at
FROM notifications
WHERE id = sqlc.arg(id);

-- name: AdminPurgeReadNotifications :exec
DELETE FROM notifications
WHERE read_at IS NOT NULL AND read_at < (NOW() - INTERVAL 24 HOUR);

-- name: SystemGetLastNotificationByMessage :one
SELECT id, users_idusers, link, message, created_at, read_at
FROM notifications
WHERE users_idusers = sqlc.arg(user_id) AND message = sqlc.arg(message)
ORDER BY id DESC LIMIT 1;

-- name: AdminListRecentNotifications :many
SELECT id, users_idusers, link, message, created_at, read_at
FROM notifications
ORDER BY id DESC LIMIT ?;
