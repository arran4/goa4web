UPDATE forumcategory SET title = ?, description = ?, parent_category_id = ?, language_id = sqlc.narg(category_language_id) WHERE forumtopic.id = ?;

-- name: GetAllForumCategoriesWithSubcategoryCount :many
SELECT c.*, COUNT(c2.id) as SubcategoryCount,
       COUNT(t.id)   as TopicCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.id = c2.parent_category_id
LEFT JOIN forumtopic t ON c.id = t.category_id
WHERE (
    c.language_id = 0
    OR c.language_id IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_id = c.language_id
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
GROUP BY c.id;

-- name: AdminCountForumCategories :one
SELECT COUNT(*)
FROM forumcategory c
WHERE (
    c.language_id = 0
    OR c.language_id IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_id = c.language_id
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
);

-- name: AdminListForumCategoriesWithCounts :many
SELECT c.*, COUNT(c2.id) AS SubcategoryCount,
       COUNT(t.id) AS TopicCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.id = c2.parent_category_id
LEFT JOIN forumtopic t ON c.id = t.category_id
WHERE (
    c.language_id = 0
    OR c.language_id IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_id = c.language_id
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
GROUP BY c.id
ORDER BY c.id
LIMIT ? OFFSET ?;

-- name: GetAllForumTopics :many
SELECT t.*
FROM forumtopic t
WHERE (
    t.language_id = 0
    OR t.language_id IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_id = t.language_id
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
GROUP BY t.id;

-- name: AdminListForumTopics :many
SELECT t.*
FROM forumtopic t
ORDER BY t.id
LIMIT ? OFFSET ?;

UPDATE forumtopic SET title = ?, description = ?, category_id = ?, language_id = sqlc.narg(topic_language_id) WHERE id = ?;

-- name: GetAllForumTopicsByCategoryIdForUserWithLastPosterName :many
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT t.*, lu.username AS last_author_username
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.last_author_id
WHERE t.category_id = sqlc.arg(category_id)
  AND (
      t.language_id = 0
      OR t.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_id = t.language_id
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
      AND ((t.handler = 'private' AND g.item_id = t.id) OR (t.handler <> 'private' AND (g.item_id = t.id OR g.item_id IS NULL)))
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
UPDATE forumtopic SET handler = sqlc.arg(handler) WHERE id = sqlc.arg(id);

-- name: AdminListTopicsWithUserGrantsNoRoles :many
SELECT t.id, t.title
FROM forumtopic t
WHERE t.handler <> 'private'
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum' AND g.item='topic' AND g.active=1
      AND g.item_id = t.id
      AND g.user_id IS NOT NULL
      AND (sqlc.arg(include_admin) OR g.user_id <> 1)
  )
  AND NOT EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum' AND g.item='topic' AND g.active=1
      AND g.item_id = t.id
      AND g.role_id IS NOT NULL
  )
ORDER BY t.id;

-- name: GetForumTopicsForUser :many
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT t.*, lu.username AS last_author_username
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.last_author_id
WHERE t.handler <> 'private'
  AND (
    t.language_id = 0
    OR t.language_id IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_id = t.language_id
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
      AND (g.item_id = t.id OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;

-- name: GetForumTopicByIdForUser :one
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT t.*, lu.username AS last_author_username
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.last_author_id
WHERE t.id = sqlc.arg(id)
  AND (
      t.language_id = 0
      OR t.language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_id = t.language_id
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
      AND ((t.handler = 'private' AND g.item_id = t.id) OR (t.handler <> 'private' AND (g.item_id = t.id OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;


-- name: GetAllForumCategories :many
SELECT f.*
FROM forumcategory f
WHERE (
    f.language_id = 0
    OR f.language_id IS NULL
    OR EXISTS (
        SELECT 1 FROM user_language ul
        WHERE ul.users_idusers = sqlc.arg(viewer_id)
          AND ul.language_id = f.language_id
    )
    OR NOT EXISTS (
        SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
);

INSERT INTO forumcategory (parent_category_id, language_id, title, description) VALUES (?, sqlc.narg(category_language_id), ?, ?);

INSERT INTO forumtopic (category_id, language_id, title, description, handler) VALUES (?, sqlc.narg(topic_language_id), ?, ?, ?);

-- name: CreateForumTopicForPoster :execlastid
INSERT INTO forumtopic (category_id, language_id, title, description, handler)
SELECT sqlc.arg(forumcategory_id), sqlc.arg(forum_lang), sqlc.arg(title), sqlc.arg(description), sqlc.arg(handler)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section=sqlc.arg(section)
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='create'
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
WHERE id = ?;

-- name: GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText :many
WITH role_ids AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT th.*, lu.username AS last_author_username, lu.idusers AS last_author_user_id, fcu.username as first_comment_username, fc.written as first_comment_written, fc.text as first_comment_text
FROM forumthread th
LEFT JOIN forumtopic t ON th.topic_id=t.id
LEFT JOIN users lu ON lu.idusers = t.last_author_id
LEFT JOIN comments fc ON th.first_comment_id=fc.idcomments
LEFT JOIN users fcu ON fcu.idusers = fc.users_idusers
WHERE th.topic_id=sqlc.arg(topic_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='forum'
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND ((t.handler = 'private' AND g.item_id = t.id) OR (t.handler <> 'private' AND (g.item_id = t.id OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY th.lastaddition DESC;

-- name: ListPrivateTopicsByUserID :many
SELECT t.id,
       t.last_author_id,
       t.category_id,
       t.language_id,
       t.title,
       t.description,
       t.threads,
       t.comments,
       t.lastaddition,
       t.handler,
       lu.username AS last_author_username
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.last_author_id
JOIN grants g ON g.item_id = t.id
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
    SELECT COUNT(id)
    FROM forumthread
    WHERE topic_id = forumtopic.id
), comments = (
    SELECT SUM(comments)
    FROM forumthread
    WHERE topic_id = forumtopic.id
), lastaddition = (
    SELECT lastaddition
    FROM forumthread
    WHERE topic_id = forumtopic.id
    ORDER BY lastaddition DESC
    LIMIT 1
), last_author_id = (
    SELECT last_author_id
    FROM forumthread
    WHERE topic_id = forumtopic.id
    ORDER BY lastaddition DESC
    LIMIT 1
);

-- name: SystemRebuildForumTopicMetaByID :exec
UPDATE forumtopic
SET threads = (
    SELECT COUNT(id)
    FROM forumthread
    WHERE topic_id = forumtopic.id
), comments = (
    SELECT SUM(comments)
    FROM forumthread
    WHERE topic_id = forumtopic.id
), lastaddition = (
    SELECT lastaddition
    FROM forumthread
    WHERE topic_id = forumtopic.id
    ORDER BY lastaddition DESC
    LIMIT 1
), last_author_id = (
    SELECT last_author_id
    FROM forumthread
    WHERE topic_id = forumtopic.id
    ORDER BY lastaddition DESC
    LIMIT 1
)
WHERE forumtopic.id = ?;

-- name: AdminDeleteForumCategory :exec
UPDATE forumcategory SET deleted_at = NOW() WHERE id = ?;

-- name: AdminDeleteForumTopic :exec
-- Removes a forum topic by ID.
UPDATE forumtopic SET deleted_at = NOW() WHERE id = ?;


-- name: GetAllForumThreadsWithTopic :many
SELECT th.*, t.title AS topic_title
FROM forumthread th
LEFT JOIN forumtopic t ON th.topic_id = t.id
ORDER BY t.id, th.lastaddition DESC;

-- name: GetForumCategoryById :one
SELECT * FROM forumcategory
WHERE id = sqlc.arg(id)
  AND (
      language_id = 0
      OR language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_id = language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  );

-- name: GetForumTopicsByCategoryId :many
SELECT * FROM forumtopic
WHERE category_id = sqlc.arg(category_id)
  AND (
      language_id = 0
      OR language_id IS NULL
      OR EXISTS (
          SELECT 1 FROM user_language ul
          WHERE ul.users_idusers = sqlc.arg(viewer_id)
            AND ul.language_id = language_id
      )
      OR NOT EXISTS (
          SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
      )
  )
ORDER BY lastaddition DESC;

-- name: ListForumcategoryPath :many
WITH RECURSIVE category_path AS (
    SELECT f.id, f.parent_category_id AS parent_id, f.title, 0 AS depth
    FROM forumcategory f
    WHERE f.id = sqlc.arg(category_id)
    UNION ALL
    SELECT c.id, c.parent_category_id, c.title, p.depth + 1
    FROM forumcategory c
    JOIN category_path p ON p.parent_id = c.id
)
SELECT category_path.id, category_path.title
FROM category_path
ORDER BY category_path.depth DESC;

