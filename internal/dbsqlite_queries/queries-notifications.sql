-- name: SystemCreateNotification :exec
INSERT INTO notifications (users_idusers, link, message)
VALUES (sqlc.arg(recipient_id), sqlc.arg(link), sqlc.arg(message));

-- name: SystemCreateNotificationForActivity :exec
INSERT INTO notifications (users_idusers, link, message, activity_key, priority, resource_type, resource_id, activity_position, read_at)
VALUES (sqlc.arg(recipient_id), sqlc.arg(link), sqlc.arg(message), sqlc.arg(activity_key), sqlc.arg(priority), sqlc.arg(resource_type), sqlc.arg(resource_id), sqlc.arg(activity_position), sqlc.arg(read_at))
ON CONFLICT(users_idusers, activity_key) DO UPDATE SET
    link = CASE WHEN excluded.priority > notifications.priority THEN excluded.link ELSE notifications.link END,
    message = CASE WHEN excluded.priority > notifications.priority THEN excluded.message ELSE notifications.message END,
    priority = MAX(notifications.priority, excluded.priority),
    resource_type = COALESCE(notifications.resource_type, excluded.resource_type),
    resource_id = COALESCE(notifications.resource_id, excluded.resource_id),
    activity_position = COALESCE(notifications.activity_position, excluded.activity_position),
    read_at = COALESCE(notifications.read_at, excluded.read_at);

-- name: SystemAcknowledgeNotificationsForResource :exec
UPDATE notifications
SET read_at = CURRENT_TIMESTAMP
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
SELECT notifications.id, notifications.users_idusers, notifications.link, notifications.message, notifications.activity_key, notifications.priority, notifications.resource_type, notifications.resource_id, notifications.activity_position, notifications.created_at, notifications.read_at
FROM notifications
WHERE users_idusers = sqlc.arg(lister_id)
ORDER BY id DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListUnreadNotificationsForLister :many
SELECT notifications.id, notifications.users_idusers, notifications.link, notifications.message, notifications.activity_key, notifications.priority, notifications.resource_type, notifications.resource_id, notifications.activity_position, notifications.created_at, notifications.read_at
FROM notifications
WHERE users_idusers = sqlc.arg(lister_id) AND read_at IS NULL
ORDER BY id DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: SetNotificationReadForLister :exec
UPDATE notifications
SET read_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(lister_id) AND read_at IS NULL;

-- name: GetNotificationIDByActivityForLister :one
SELECT id FROM notifications
WHERE users_idusers = sqlc.arg(lister_id) AND activity_key = sqlc.arg(activity_key);

-- name: SetNotificationsReadForListerBatch :exec
UPDATE notifications
SET read_at = CURRENT_TIMESTAMP
WHERE users_idusers = sqlc.arg(lister_id)
  AND id IN (sqlc.slice('ids'))
  AND read_at IS NULL;

-- name: SetNotificationUnreadForLister :exec
UPDATE notifications
SET read_at = NULL
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(lister_id);

-- name: DeleteNotificationForLister :exec
DELETE FROM notifications
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(lister_id);

-- name: GetNotificationForLister :one
SELECT notifications.id, notifications.users_idusers, notifications.link, notifications.message, notifications.activity_key, notifications.priority, notifications.resource_type, notifications.resource_id, notifications.activity_position, notifications.created_at, notifications.read_at
FROM notifications
WHERE id = sqlc.arg(id) AND users_idusers = sqlc.arg(lister_id);

-- name: AdminGetNotification :one
SELECT notifications.id, notifications.users_idusers, notifications.link, notifications.message, notifications.activity_key, notifications.priority, notifications.resource_type, notifications.resource_id, notifications.activity_position, notifications.created_at, notifications.read_at
FROM notifications
WHERE id = sqlc.arg(id);

-- name: AdminMarkNotificationRead :exec
UPDATE notifications SET read_at = CURRENT_TIMESTAMP WHERE id = sqlc.arg(id);

-- name: AdminMarkNotificationUnread :exec
UPDATE notifications SET read_at = NULL WHERE id = sqlc.arg(id);

-- name: AdminDeleteNotification :exec
DELETE FROM notifications WHERE id = sqlc.arg(id);

-- name: AdminPurgeReadNotifications :exec
DELETE FROM notifications
WHERE read_at IS NOT NULL AND read_at < (datetime('now', '-24 hours'));

-- name: SystemGetLastNotificationForRecipientByMessage :one
SELECT notifications.id, notifications.users_idusers, notifications.link, notifications.message, notifications.activity_key, notifications.priority, notifications.resource_type, notifications.resource_id, notifications.activity_position, notifications.created_at, notifications.read_at
FROM notifications
WHERE users_idusers = sqlc.arg(recipient_id) AND message = sqlc.arg(message)
ORDER BY id DESC LIMIT 1;

-- name: AdminListRecentNotifications :many
SELECT notifications.id, notifications.users_idusers, notifications.link, notifications.message, notifications.activity_key, notifications.priority, notifications.resource_type, notifications.resource_id, notifications.activity_position, notifications.created_at, notifications.read_at
FROM notifications
ORDER BY id DESC LIMIT sqlc.arg(limit);

-- name: AdminDeleteNotificationsByMessage :exec
DELETE FROM notifications
WHERE message LIKE ?;

-- name: ListVisibleNotificationEntriesForLister :many
WITH viewer AS (
    SELECT u.idusers AS viewer_id FROM users u WHERE u.idusers = sqlc.arg(lister_id) AND u.deleted_at IS NULL
),
viewer_roles AS (
    SELECT ur.role_id AS id FROM user_roles ur JOIN viewer v ON v.viewer_id = ur.users_idusers
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
),
viewer_languages AS (
    SELECT ul.language_id AS id FROM user_language ul JOIN viewer v ON v.viewer_id = ul.users_idusers
),
visible AS (
    SELECT n.id, n.users_idusers, n.link, n.message, n.activity_key, n.priority,
           n.resource_type, n.resource_id, n.activity_position, n.created_at, n.read_at,
           COALESCE(t.title, '') AS topic_title,
           COALESCE(author.username, '') AS reply_author,
           CASE WHEN n.resource_type = 'thread' AND n.activity_key LIKE 'reply:comment:%' THEN 1 ELSE 0 END AS forum_reply,
           CAST(CASE WHEN sqlc.arg(group_unread) = 1 AND n.read_at IS NULL
                     AND n.resource_type = 'thread' AND n.activity_key LIKE 'reply:comment:%'
                THEN n.resource_id ELSE NULL END AS INTEGER) AS reply_group
    FROM notifications n
    JOIN viewer v ON v.viewer_id = n.users_idusers
    LEFT JOIN forumthread th ON n.resource_type = 'thread' AND th.idforumthread = n.resource_id
    LEFT JOIN forumtopic t ON t.idforumtopic = th.forumtopic_idforumtopic
    LEFT JOIN comments opener ON opener.idcomments = th.firstpost
    LEFT JOIN comments reply ON n.activity_key LIKE 'reply:comment:%'
        AND reply.idcomments = (n.activity_position >> 32) AND reply.forumthread_id = th.idforumthread
    LEFT JOIN users author ON author.idusers = reply.users_idusers
    WHERE (sqlc.arg(include_read) OR n.read_at IS NULL)
      AND (sqlc.arg(notification_id) = 0 OR n.id = sqlc.arg(notification_id))
      AND (
          sqlc.arg(delivery_id) = 0 OR EXISTS (
              SELECT 1 FROM digest_delivery_notifications item
              WHERE item.notification_id = n.id AND item.digest_delivery_id = sqlc.arg(delivery_id)
          )
      )
      AND (
          (COALESCE(n.resource_type, '') <> 'thread'
              AND LOWER(COALESCE(n.link, '')) NOT LIKE '%/private/%')
          OR (n.resource_type = 'thread'
              AND (t.language_id IS NULL OR t.language_id = 0 OR t.language_id IN (SELECT id FROM viewer_languages)
                  OR NOT EXISTS (SELECT 1 FROM viewer_languages))
              AND (opener.language_id IS NULL OR opener.language_id = 0 OR opener.language_id IN (SELECT id FROM viewer_languages)
                  OR NOT EXISTS (SELECT 1 FROM viewer_languages))
              AND (
                  COALESCE(n.activity_key, '') NOT LIKE 'reply:comment:%'
                  OR (reply.idcomments IS NOT NULL AND reply.deleted_at IS NULL
                      AND (reply.language_id IS NULL OR reply.language_id = 0 OR reply.language_id IN (SELECT id FROM viewer_languages)
                          OR NOT EXISTS (SELECT 1 FROM viewer_languages)))
              )
              AND EXISTS (
                  SELECT 1 FROM grants g
                  WHERE g.section = CASE WHEN t.handler = 'private' THEN 'privateforum' ELSE 'forum' END
                    AND (g.item = 'topic' OR g.item IS NULL)
                    AND g.action = 'view' AND g.active = 1
                    AND (g.item_id = t.idforumtopic OR (t.handler <> 'private' AND g.item_id IS NULL))
                    AND (g.user_id = v.viewer_id OR g.user_id IS NULL)
                    AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM viewer_roles))
              )
              AND (t.handler <> 'private' OR EXISTS (
                  SELECT 1 FROM grants g
                  WHERE g.section = 'privateforum_thread' AND g.item = 'thread'
                    AND g.action = 'view' AND g.active = 1 AND g.item_id = th.idforumthread
                    AND (g.user_id = v.viewer_id OR g.user_id IS NULL)
                    AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM viewer_roles))
              ))
          )
      )
),
reply_groups AS (
    SELECT v.reply_group, COUNT(*) AS reply_count, MAX(v.id) AS latest_notification_id,
           json_group_array(v.id) AS snapshot
    FROM visible v
    WHERE v.reply_group IS NOT NULL
    GROUP BY v.reply_group
),
ranked AS (
    SELECT v.id, v.users_idusers, v.link, v.message, v.activity_key, v.priority,
           v.resource_type, v.resource_id, v.activity_position, v.created_at, v.read_at,
           v.topic_title, v.reply_author, v.forum_reply, v.reply_group, ROW_NUMBER() OVER (
        PARTITION BY v.reply_group, CASE WHEN v.reply_group IS NULL THEN v.id ELSE 0 END
        ORDER BY v.activity_position DESC, v.id DESC
    ) AS display_order
    FROM visible v
)
SELECT ranked.id, ranked.users_idusers, ranked.link, ranked.message, ranked.activity_key,
       ranked.priority, ranked.resource_type, ranked.resource_id, ranked.activity_position,
       ranked.created_at, ranked.read_at, ranked.topic_title, ranked.reply_author, ranked.forum_reply,
       CAST(COALESCE(reply_groups.reply_count, 1) AS INTEGER) AS reply_count,
       CAST(COALESCE(reply_groups.snapshot, '') AS TEXT) AS snapshot,
       CAST(COUNT(*) OVER () AS INTEGER) AS total_entries
FROM ranked
LEFT JOIN reply_groups ON reply_groups.reply_group = ranked.reply_group
WHERE ranked.display_order = 1
ORDER BY COALESCE(reply_groups.latest_notification_id, ranked.id) DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);
