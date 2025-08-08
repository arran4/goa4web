-- name: AdminUpdateForumCategory :exec
UPDATE forumcategory SET title = ?, description = ?, forumcategory_idforumcategory = ?, language_idlanguage = ? WHERE idforumcategory = ?;

-- name: GetAllForumCategoriesWithSubcategoryCount :many
SELECT c.*, COUNT(c2.idforumcategory) as SubcategoryCount,
       COUNT(t.idforumtopic)   as TopicCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.idforumcategory = c2.forumcategory_idforumcategory
LEFT JOIN forumtopic t ON c.idforumcategory = t.forumcategory_idforumcategory
WHERE (
    c.language_idlanguage = 0
    OR c.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_idlanguage = c.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
GROUP BY c.idforumcategory;

-- name: AdminCountForumCategories :one
SELECT COUNT(*)
FROM forumcategory c
WHERE (
    c.language_idlanguage = 0
    OR c.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_idlanguage = c.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
);

-- name: AdminListForumCategoriesWithCounts :many
SELECT c.*, COUNT(c2.idforumcategory) AS SubcategoryCount,
       COUNT(t.idforumtopic) AS TopicCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.idforumcategory = c2.forumcategory_idforumcategory
LEFT JOIN forumtopic t ON c.idforumcategory = t.forumcategory_idforumcategory
WHERE (
    c.language_idlanguage = 0
    OR c.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_idlanguage = c.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
GROUP BY c.idforumcategory
ORDER BY c.idforumcategory
LIMIT ? OFFSET ?;

-- name: GetAllForumTopics :many
SELECT t.*
FROM forumtopic t
WHERE (
    t.language_idlanguage = 0
    OR t.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_idlanguage = t.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
GROUP BY t.idforumtopic;

-- name: AdminListForumTopics :many
SELECT t.*
FROM forumtopic t
ORDER BY t.idforumtopic
LIMIT ? OFFSET ?;

-- name: AdminUpdateForumTopic :exec
UPDATE forumtopic SET title = ?, description = ?, forumcategory_idforumcategory = ?, language_idlanguage = ? WHERE idforumtopic = ?;

-- name: GetAllForumTopicsByCategoryIdForUserWithLastPosterName :many
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.forumcategory_idforumcategory = sqlc.arg(category_id)
  AND (
      t.language_idlanguage = 0
      OR t.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = t.language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND ((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;

-- name: ListPrivateTopicParticipantsByTopicIDForUser :many
SELECT u.idusers, u.username
FROM grants g
JOIN users u ON u.idusers = g.user_id
WHERE g.section = 'forum'
  AND g.item = 'topic'
  AND g.action = 'view'
  AND g.active = 1
  AND g.user_id IS NOT NULL
  AND g.item_id = sqlc.arg(topic_id)
  AND EXISTS (
      SELECT 1 FROM grants pg
      WHERE pg.section='forum'
        AND pg.item='topic'
        AND pg.action='view'
        AND pg.active=1
        AND pg.item_id = g.item_id
        AND pg.user_id = sqlc.arg(viewer_id)
  );

-- name: SystemSetForumTopicHandlerByID :exec
UPDATE forumtopic SET handler = sqlc.arg(handler) WHERE idforumtopic = sqlc.arg(id);

-- name: AdminListTopicsWithUserGrantsNoRoles :many
SELECT t.idforumtopic, t.title
FROM forumtopic t
WHERE t.handler <> 'private'
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum' AND g.item='topic' AND g.active=1
      AND g.item_id = t.idforumtopic
      AND g.user_id IS NOT NULL
      AND (sqlc.arg(include_admin) OR g.user_id <> 1)
  )
  AND NOT EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum' AND g.item='topic' AND g.active=1
      AND g.item_id = t.idforumtopic
      AND g.role_id IS NOT NULL
  )
ORDER BY t.idforumtopic;

-- name: GetForumTopicsForUser :many
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.handler <> 'private'
  AND (
    t.language_idlanguage = 0
    OR t.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_idlanguage = t.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
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

-- name: GetForumTopicByIdForUser :one
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.idforumtopic = sqlc.arg(idforumtopic)
  AND (
      t.language_idlanguage = 0
      OR t.language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = t.language_idlanguage
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
      AND ((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;


-- name: GetAllForumCategories :many
SELECT f.*
FROM forumcategory f
WHERE (
    f.language_idlanguage = 0
    OR f.language_idlanguage IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_idlanguage = f.language_idlanguage
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
);

-- name: AdminCreateForumCategory :exec
INSERT INTO forumcategory (forumcategory_idforumcategory, language_idlanguage, title, description) VALUES (?, ?, ?, ?);

-- name: SystemCreateForumTopic :execlastid
INSERT INTO forumtopic (forumcategory_idforumcategory, language_idlanguage, title, description, handler) VALUES (?, ?, ?, ?, ?);

-- name: CreateForumTopicForPoster :execlastid
INSERT INTO forumtopic (forumcategory_idforumcategory, language_idlanguage, title, description, handler)
SELECT sqlc.arg(forumcategory_id), sqlc.narg(language_id), sqlc.arg(title), sqlc.arg(description), sqlc.arg(handler)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section=sqlc.arg(section)
      AND (g.item='category' OR g.item IS NULL)
      AND g.action='post'
      AND g.active=1
      AND (g.item_id = sqlc.arg(grant_category_id) OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(poster_id)
      ))
  );

-- name: SystemGetForumTopicByTitle :one
SELECT *
FROM forumtopic
WHERE title=?;

-- name: GetForumTopicById :one
SELECT *
FROM forumtopic
WHERE idforumtopic = ?;

-- name: GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText :many
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
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
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND ((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY th.lastaddition DESC;

-- name: ListPrivateTopicsByUserID :many
SELECT t.idforumtopic,
       t.lastposter,
       t.forumcategory_idforumcategory,
       t.language_idlanguage,
       t.title,
       t.description,
       t.threads,
       t.comments,
       t.lastaddition,
       t.handler,
       lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
JOIN grants g ON g.item_id = t.idforumtopic
WHERE t.handler = 'private'
  AND g.section = 'forum'
  AND g.item = 'topic'
  AND g.action = 'view'
  AND g.active = 1
  AND g.user_id = sqlc.arg(user_id)
ORDER BY t.lastaddition DESC;

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
SELECT * FROM forumcategory
WHERE idforumcategory = sqlc.arg(idforumcategory)
  AND (
      language_idlanguage = 0
      OR language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  );

-- name: GetForumTopicsByCategoryId :many
SELECT * FROM forumtopic
WHERE forumcategory_idforumcategory = sqlc.arg(category_id)
  AND (
      language_idlanguage = 0
      OR language_idlanguage IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_idlanguage = language_idlanguage
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
ORDER BY lastaddition DESC;

-- name: ListForumcategoryPath :many
WITH RECURSIVE category_path AS (
    SELECT f.idforumcategory, f.forumcategory_idforumcategory AS parent_id, f.title, 0 AS depth
    FROM forumcategory f
    WHERE f.idforumcategory = sqlc.arg(category_id)
    UNION ALL
    SELECT c.idforumcategory, c.forumcategory_idforumcategory, c.title, p.depth + 1
    FROM forumcategory c
    JOIN category_path p ON p.parent_id = c.idforumcategory
)
SELECT category_path.idforumcategory, category_path.title
FROM category_path
ORDER BY category_path.depth DESC;
