-- name: AdminUpdateForumCategory :exec
UPDATE forumcategory SET title = ?, description = ?, forumcategory_idforumcategory = ? WHERE idforumcategory = ?;

-- name: GetAllForumCategoriesWithSubcategoryCount :many
SELECT c.*, COUNT(c2.idforumcategory) as SubcategoryCount,
       COUNT(t.idforumtopic)   as TopicCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.idforumcategory = c2.forumcategory_idforumcategory
LEFT JOIN forumtopic t ON c.idforumcategory = t.forumcategory_idforumcategory
GROUP BY c.idforumcategory;

-- name: GetAllForumTopics :many
SELECT t.*
FROM forumtopic t
GROUP BY t.idforumtopic;

-- name: AdminUpdateForumTopic :exec
UPDATE forumtopic SET title = ?, description = ?, forumcategory_idforumcategory = ? WHERE idforumtopic = ?;

-- name: GetAllForumTopicsByCategoryIdForUserWithLastPosterName :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section='role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.forumcategory_idforumcategory = sqlc.arg(category_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;

-- name: GetAllForumTopicsForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section='role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;

-- name: GetForumTopicByIdForUser :one
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section='role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.idforumtopic = sqlc.arg(idforumtopic)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND g.item='topic'
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;


-- name: GetAllForumCategories :many
SELECT f.*
FROM forumcategory f;

-- name: AdminCreateForumCategory :exec
INSERT INTO forumcategory (forumcategory_idforumcategory, title, description) VALUES (?, ?, ?);

-- name: SystemCreateForumTopic :execlastid
INSERT INTO forumtopic (forumcategory_idforumcategory, title, description) VALUES (?, ?, ?);

-- name: SystemGetForumTopicByTitle :one
SELECT *
FROM forumtopic
WHERE title=?;

-- name: GetForumTopicById :one
SELECT *
FROM forumtopic
WHERE idforumtopic = ?;

-- name: GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section='role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT th.*, lu.username AS lastposterusername, lu.idusers AS lastposterid, fcu.username as firstpostusername, fc.written as firstpostwritten, fc.text as firstposttext
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users lu ON lu.idusers = t.lastposter
LEFT JOIN comments fc ON th.firstpost=fc.idcomments
LEFT JOIN users fcu ON fcu.idusers = fc.users_idusers
WHERE th.forumtopic_idforumtopic=sqlc.arg(topic_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND g.item='topic'
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY th.lastaddition DESC;

-- name: AdminRebuildAllForumTopicMetaColumns :exec
UPDATE forumtopic
SET threads = (
    SELECT COUNT(idforumthread)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), comments = (
    SELECT SUM(comments)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), lastaddition = (
    SELECT lastaddition
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
), lastposter = (
    SELECT lastposter
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
);

-- name: SystemRebuildForumTopicMetaByID :exec
UPDATE forumtopic
SET threads = (
    SELECT COUNT(idforumthread)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), comments = (
    SELECT SUM(comments)
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
), lastaddition = (
    SELECT lastaddition
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
), lastposter = (
    SELECT lastposter
    FROM forumthread
    WHERE forumtopic_idforumtopic = idforumtopic
    ORDER BY lastaddition DESC
    LIMIT 1
)
WHERE idforumtopic = ?;

-- name: AdminDeleteForumCategory :exec
UPDATE forumcategory SET deleted_at = NOW() WHERE idforumcategory = ?;

-- name: AdminDeleteForumTopic :exec
-- Removes a forum topic by ID.
UPDATE forumtopic SET deleted_at = NOW() WHERE idforumtopic = ?;


-- name: GetAllForumThreadsWithTopic :many
SELECT th.*, t.title AS topic_title
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
ORDER BY t.idforumtopic, th.lastaddition DESC;

-- name: GetForumCategoryById :one
SELECT * FROM forumcategory WHERE idforumcategory = ?;

-- name: GetForumTopicsByCategoryId :many
SELECT * FROM forumtopic WHERE forumcategory_idforumcategory = ? ORDER BY lastaddition DESC;
