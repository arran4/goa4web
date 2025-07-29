-- name: RecalculateAllForumThreadMetaData :exec
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

-- name: RecalculateForumThreadByIdMetaData :exec
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

-- name: GetThreadLastPosterAndPerms :one
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section='role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT th.*, lu.username AS LastPosterUsername
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE th.idforumthread=sqlc.arg(thread_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND g.item='topic'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = t.idforumtopic
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;

-- name: MakeThread :execlastid
INSERT INTO forumthread (forumtopic_idforumtopic) VALUES (?);


-- name: GetForumTopicIdByThreadId :one
SELECT forumtopic_idforumtopic FROM forumthread WHERE idforumthread = ?;

-- name: DeleteForumThread :exec
UPDATE forumthread SET deleted_at = NOW() WHERE idforumthread = ?;


-- name: GetThreadsStartedByUser :many
SELECT th.*
FROM forumthread th
JOIN comments c ON th.firstpost = c.idcomments
WHERE c.users_idusers = ?
ORDER BY th.lastaddition DESC;

-- name: GetThreadsStartedByUserWithTopic :many
SELECT th.*, t.title AS topic_title, fc.idforumcategory AS category_id, fc.title AS category_title
FROM forumthread th
JOIN comments c ON th.firstpost = c.idcomments
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
LEFT JOIN forumcategory fc ON t.forumcategory_idforumcategory = fc.idforumcategory
WHERE c.users_idusers = ?
ORDER BY th.lastaddition DESC;
