-- name: AdminRecalculateAllForumThreadMetaData :exec
UPDATE forumthread
SET lastaddition = (
    SELECT written
    FROM comments
    WHERE forumthread_id = idforumthread
    ORDER BY written DESC
    LIMIT 1
), comments = (
    SELECT COUNT(users_idusers) - 1
    FROM comments
    WHERE forumthread_id = idforumthread
), lastposter = (
    SELECT users_idusers
    FROM comments
    WHERE forumthread_id = idforumthread
    ORDER BY written DESC
    LIMIT 1
), firstpost = (
    SELECT idcomments
    FROM comments
    WHERE forumthread_id = idforumthread
    LIMIT 1
);

-- name: AdminRecalculateForumThreadByIdMetaData :exec
UPDATE forumthread
SET lastaddition = (
    SELECT written
    FROM comments
    WHERE forumthread_id = idforumthread
    ORDER BY written DESC
    LIMIT 1
), comments = (
    SELECT COUNT(users_idusers) - 1
    FROM comments
    WHERE forumthread_id = idforumthread
), lastposter = (
    SELECT users_idusers
    FROM comments
    WHERE forumthread_id = idforumthread
    ORDER BY written DESC
    LIMIT 1
), firstpost = (
    SELECT idcomments
    FROM comments
    WHERE forumthread_id = idforumthread
    LIMIT 1
)
WHERE idforumthread = ?;

-- name: AdminListForumThreads :many
SELECT
    t.idforumthread,
    t.forumtopic_idforumtopic as idforumtopic,
    SUBSTRING(c.text, 1, 100) AS title,
    c.written as created_at,
    c.users_idusers as created_by,
    t.lastposter as last_post_by,
    t.lastaddition as last_post_at,
    t.comments as post_count,
    ft.title as topic_title,
    ft.handler as topic_handler
FROM
    forumthread t
JOIN
    forumtopic ft ON t.forumtopic_idforumtopic = ft.idforumtopic
JOIN
    comments c ON t.firstpost = c.idcomments
ORDER BY t.idforumthread
LIMIT ? OFFSET ?;

-- name: GetThreadLastPosterAndPerms :one
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT th.*, lu.username AS LastPosterUsername
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users lu ON lu.idusers = t.lastposter
LEFT JOIN comments fc ON th.firstpost = fc.idcomments
WHERE th.idforumthread=sqlc.arg(thread_id)
  AND (
      fc.language_id = 0
      OR fc.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_id = fc.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE (g.section='forum' OR g.section='privateforum')
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  ) AND (t.handler IS NULL OR t.handler != 'private' OR EXISTS (
    SELECT 1 FROM grants g
    WHERE (g.section='privateforum')
      AND g.item='thread'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = th.idforumthread
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  ))
ORDER BY t.lastaddition DESC;

-- name: SystemCreateThread :execlastid
INSERT INTO forumthread (forumtopic_idforumtopic) VALUES (?);


-- name: GetForumTopicIdByThreadId :one
SELECT forumtopic_idforumtopic FROM forumthread WHERE idforumthread = ?;

-- name: AdminDeleteForumThread :exec
DELETE FROM forumthread WHERE idforumthread = ?;

-- name: AdminListForumThreadGrantsByThreadID :many
SELECT
    g.id,
    g.section,
    g.action,
    r.name AS role_name,
    u.username
FROM
    grants g
LEFT JOIN
    roles r ON g.role_id = r.id
LEFT JOIN
    users u ON g.user_id = u.idusers
WHERE
    g.section = 'forum'
    AND g.item = 'thread'
    AND g.item_id = ?;


-- name: AdminGetThreadsStartedByUser :many
SELECT th.*
FROM forumthread th
JOIN comments c ON th.firstpost = c.idcomments
WHERE c.users_idusers = ?
ORDER BY th.lastaddition DESC;

-- name: AdminGetThreadsStartedByUserWithTopic :many
SELECT th.*, t.title AS topic_title, fc.idforumcategory AS category_id, fc.title AS category_title
FROM forumthread th
JOIN comments c ON th.firstpost = c.idcomments
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
LEFT JOIN forumcategory fc ON t.forumcategory_idforumcategory = fc.idforumcategory
WHERE c.users_idusers = ?
ORDER BY th.lastaddition DESC;

-- name: GetThreadBySectionThreadIDForReplier :one
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(replier_id)
)
SELECT th.*
FROM forumthread th
LEFT JOIN comments fc ON th.firstpost = fc.idcomments
WHERE th.idforumthread = sqlc.arg(thread_id)
  AND (
      fc.language_id = 0
      OR fc.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(replier_id)
            AND ul.language_id = fc.language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(replier_id)
      )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = sqlc.arg(section)
      AND (g.item = sqlc.arg(item_type) OR g.item IS NULL)
      AND g.action = 'reply'
      AND g.active = 1
      AND (g.item_id = sqlc.arg(item_id) OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(replier_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );
