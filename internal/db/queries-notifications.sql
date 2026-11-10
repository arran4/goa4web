-- name: SystemCreateNotification :exec
INSERT INTO notifications (users_idusers, link, message)
VALUES (sqlc.arg(recipient_id), sqlc.arg(link), sqlc.arg(message));

-- name: SystemCreateNotificationForActivity :exec
INSERT INTO notifications (users_idusers, link, message, activity_key, priority, resource_type, resource_id, activity_position, read_at)
VALUES (sqlc.arg(recipient_id), sqlc.arg(link), sqlc.arg(message), sqlc.arg(activity_key), sqlc.arg(priority), sqlc.arg(resource_type), sqlc.arg(resource_id), sqlc.arg(activity_position), sqlc.arg(read_at))
ON DUPLICATE KEY UPDATE
    link = IF(VALUES(priority) > priority, VALUES(link), link),
    message = IF(VALUES(priority) > priority, VALUES(message), message),
    priority = GREATEST(priority, VALUES(priority)),
    resource_type = COALESCE(resource_type, VALUES(resource_type)),
    resource_id = COALESCE(resource_id, VALUES(resource_id)),
    activity_position = COALESCE(activity_position, VALUES(activity_position));

-- name: SystemAcknowledgeNotificationsForResource :exec
UPDATE notifications
SET read_at = NOW()
WHERE users_idusers = sqlc.arg(recipient_id)
  AND resource_type = sqlc.arg(resource_type)
  AND resource_id = sqlc.arg(resource_id)
  AND activity_position <= sqlc.arg(activity_position)
  AND read_at IS NULL;

-- GetUnreadNotificationCountForLister returns the number of unread notifications for a
-- lister.
-- Parameters:
--   lister_id - ID of the lister to count notifications for
-- name: GetUnreadNotificationCountForLister :one
SELECT COUNT(*) FROM notifications
WHERE users_idusers = sqlc.arg(lister_id) AND read_at IS NULL;

-- GetNotificationCountForLister returns the total number of notifications for a
-- lister.
-- Parameters:
--   lister_id - ID of the lister to count notifications for
-- name: GetNotificationCountForLister :one
SELECT COUNT(*) FROM notifications
WHERE users_idusers = sqlc.arg(lister_id);

-- name: ListNotificationsForLister :many
SELECT notifications.*
FROM notifications
WHERE users_idusers = sqlc.arg(lister_id)
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: ListUnreadNotificationsForLister :many
SELECT notifications.*
FROM notifications
WHERE users_idusers = sqlc.arg(lister_id) AND read_at IS NULL
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SetNotificationReadForLister :exec
UPDATE notifications
SET read_at = NOW()
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(lister_id);

-- name: SetNotificationsReadForListerBatch :exec
UPDATE notifications
SET read_at = NOW()
WHERE users_idusers = sqlc.arg(lister_id)
  AND id IN (sqlc.slice('ids'));

-- name: SetNotificationUnreadForLister :exec
UPDATE notifications
SET read_at = NULL
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(lister_id);

-- name: DeleteNotificationForLister :exec
DELETE FROM notifications
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(lister_id);

-- name: GetNotificationForLister :one
SELECT notifications.*
FROM notifications
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(lister_id);

-- name: AdminGetNotification :one
SELECT notifications.*
FROM notifications
WHERE id = sqlc.arg(id);

-- name: AdminMarkNotificationRead :exec
UPDATE notifications SET read_at = NOW() WHERE id = sqlc.arg(id);

-- name: AdminMarkNotificationUnread :exec
UPDATE notifications SET read_at = NULL WHERE id = sqlc.arg(id);

-- name: AdminDeleteNotification :exec
DELETE FROM notifications WHERE id = sqlc.arg(id);

-- name: AdminPurgeReadNotifications :exec
DELETE FROM notifications
WHERE read_at IS NOT NULL AND read_at < (NOW() - INTERVAL 24 HOUR);

-- name: SystemGetLastNotificationForRecipientByMessage :one
SELECT notifications.*
FROM notifications
WHERE users_idusers = sqlc.arg(recipient_id) AND message = sqlc.arg(message)
ORDER BY id DESC LIMIT 1;

-- name: AdminListRecentNotifications :many
SELECT notifications.*
FROM notifications
ORDER BY id DESC LIMIT ?;

-- name: AdminDeleteNotificationsByMessage :exec
DELETE FROM notifications
WHERE message LIKE ?;
