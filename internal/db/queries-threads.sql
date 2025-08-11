-- name: AdminRecalculateAllForumThreadMetaData :exec
UPDATE forumthread
SET lastaddition = (
    SELECT written
    FROM comments
    WHERE forumthread_id = id
    ORDER BY written DESC
    LIMIT 1
), comments = (
    SELECT COUNT(users_idusers) - 1
    FROM comments
    WHERE forumthread_id = id
), last_author_id = (
    SELECT users_idusers
    FROM comments
    WHERE forumthread_id = id
    ORDER BY written DESC
    LIMIT 1
), first_comment_id = (
    SELECT idcomments
    FROM comments
    WHERE forumthread_id = id
    LIMIT 1
);

-- name: AdminRecalculateForumThreadByIdMetaData :exec
UPDATE forumthread
SET lastaddition = (
    SELECT written
    FROM comments
    WHERE forumthread_id = id
    ORDER BY written DESC
    LIMIT 1
), comments = (
    SELECT COUNT(users_idusers) - 1
    FROM comments
    WHERE forumthread_id = id
), last_author_id = (
    SELECT users_idusers
    FROM comments
    WHERE forumthread_id = id
    ORDER BY written DESC
    LIMIT 1
), first_comment_id = (
    SELECT idcomments
    FROM comments
    WHERE forumthread_id = id
    LIMIT 1
)
WHERE id = ?;

-- name: GetThreadLastPosterAndPerms :one
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT th.*, lu.username AS LastAuthorUsername
FROM forumthread th
LEFT JOIN forumtopic t ON th.topic_id=t.id
LEFT JOIN users lu ON lu.idusers = t.last_author_id
LEFT JOIN comments fc ON th.first_comment_id = fc.idcomments
WHERE th.id=sqlc.arg(thread_id)
  AND (
      fc.language_idlanguage = 0
      OR fc.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = fc.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = t.id OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;

-- name: SystemCreateThread :execlastid
INSERT INTO forumthread (topic_id) VALUES (?);


-- name: GetForumTopicIdByThreadId :one
SELECT topic_id FROM forumthread WHERE id = ?;

-- name: AdminDeleteForumThread :exec
UPDATE forumthread SET deleted_at = NOW() WHERE id = ?;


-- name: AdminGetThreadsStartedByUser :many
SELECT th.*
FROM forumthread th
JOIN comments c ON th.first_comment_id = c.idcomments
WHERE c.users_idusers = ?
ORDER BY th.lastaddition DESC;

-- name: AdminGetThreadsStartedByUserWithTopic :many
SELECT th.*, t.title AS topic_title, fc.id AS category_id, fc.title AS category_title
FROM forumthread th
JOIN comments c ON th.first_comment_id = c.idcomments
LEFT JOIN forumtopic t ON th.topic_id = t.id
LEFT JOIN forumcategory fc ON t.category_id = fc.id
WHERE c.users_idusers = ?
ORDER BY th.lastaddition DESC;

-- name: GetThreadBySectionThreadIDForReplier :one
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(replier_id)
)
SELECT th.*
FROM forumthread th
LEFT JOIN comments fc ON th.first_comment_id = fc.idcomments
WHERE th.id = sqlc.arg(thread_id)
  AND (
      fc.language_idlanguage = 0
      OR fc.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(replier_id)
            AND ul.language_idlanguage = fc.language_idlanguage
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
