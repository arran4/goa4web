#!/bin/bash
sed -i '/-- name: ListUnreadPrivateThreadsForUser :many/,$d' internal/db/queries-forum.sql
cat << 'SQL_EOF' >> internal/db/queries-forum.sql

-- name: ListUnreadPrivateThreadsForUser :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(grantee_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT th.idforumthread,
       th.forumtopic_idforumtopic as topic_id,
       t.title as topic_title,
       th.lastaddition,
       th.lastposter,
       lu.username AS lastposterusername
FROM forumthread th
JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
JOIN comments c ON th.firstpost = c.idcomments
LEFT JOIN users lu ON lu.idusers = th.lastposter
WHERE t.handler = 'private'
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'privateforum'
      AND g.item = 'topic'
      AND g.action = 'see'
      AND g.active = 1
      AND g.item_id = t.idforumtopic
      AND (g.user_id = sqlc.arg(grant_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
  AND (
      -- If thread has an unread label (and invert=false), it's unread
      EXISTS (
          SELECT 1 FROM content_private_labels cpl
          WHERE cpl.item = 'thread'
            AND cpl.item_id = th.idforumthread
            AND cpl.user_id = sqlc.arg(grantee_id)
            AND cpl.label = 'unread'
            AND cpl.invert = false
      )
      OR
      (
          -- Otherwise, if it's not marked as 'not unread'
          NOT EXISTS (
              SELECT 1 FROM content_private_labels cpl
              WHERE cpl.item = 'thread'
                AND cpl.item_id = th.idforumthread
                AND cpl.user_id = sqlc.arg(grantee_id)
                AND cpl.label = 'unread'
                AND cpl.invert = true
          )
          AND (
              -- And it's either not authored by user OR has a 'new' label explicitly
              c.users_idusers != sqlc.arg(grantee_id)
              OR EXISTS (
                  SELECT 1 FROM content_private_labels cpl
                  WHERE cpl.item = 'thread'
                    AND cpl.item_id = th.idforumthread
                    AND cpl.user_id = sqlc.arg(grantee_id)
                    AND cpl.label = 'new'
                    AND cpl.invert = false
              )
          )
      )
  )
ORDER BY th.lastaddition DESC
LIMIT ? OFFSET ?;
SQL_EOF
