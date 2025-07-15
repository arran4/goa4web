-- name: InsertNotification :exec
INSERT INTO notifications (users_idusers, link, message)
VALUES (?, ?, ?);

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications
WHERE users_idusers = ? AND read_at IS NULL;

-- name: GetUnreadNotifications :many
SELECT id, users_idusers, link, message, created_at, read_at
FROM notifications
WHERE users_idusers = ? AND read_at IS NULL
ORDER BY id DESC;

-- name: MarkNotificationRead :exec
UPDATE notifications SET read_at = NOW() WHERE id = ?;

-- name: PurgeReadNotifications :exec
DELETE FROM notifications
WHERE read_at IS NOT NULL AND read_at < (NOW() - INTERVAL 24 HOUR);

-- name: LastNotificationByMessage :one
SELECT id, users_idusers, link, message, created_at, read_at
FROM notifications
WHERE users_idusers = ? AND message = ?
ORDER BY id DESC LIMIT 1;

-- name: RecentNotifications :many
SELECT id, users_idusers, link, message, created_at, read_at
FROM notifications
ORDER BY id DESC LIMIT ?;
