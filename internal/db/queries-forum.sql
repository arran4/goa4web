-- name: AdminUpdateForumCategory :exec
UPDATE forumcategory
SET title = sqlc.arg(title),
    description = sqlc.arg(description),
    forumcategory_idforumcategory = sqlc.arg(parent_id),
    language_id = sqlc.narg(language_id)
WHERE idforumcategory = sqlc.arg(idforumcategory);

-- name: GetAllForumCategoriesWithSubcategoryCount :many
SELECT c.*, COUNT(c2.idforumcategory) as SubcategoryCount,
       COUNT(t.idforumtopic)   as TopicCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.idforumcategory = c2.forumcategory_idforumcategory
LEFT JOIN forumtopic t ON c.idforumcategory = t.forumcategory_idforumcategory
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
GROUP BY c.idforumcategory;

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
SELECT c.*, COUNT(c2.idforumcategory) AS SubcategoryCount,
       COUNT(t.idforumtopic) AS TopicCount
FROM forumcategory c
LEFT JOIN forumcategory c2 ON c.idforumcategory = c2.forumcategory_idforumcategory
LEFT JOIN forumtopic t ON c.idforumcategory = t.forumcategory_idforumcategory
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
GROUP BY c.idforumcategory
ORDER BY c.idforumcategory
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
GROUP BY t.idforumtopic;

-- name: AdminListForumTopics :many
SELECT t.*
FROM forumtopic t
ORDER BY t.idforumtopic
LIMIT ? OFFSET ?;

-- name: AdminListForumTopicGrantsByTopicID :many
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
    AND (g.item = 'topic' OR g.item IS NULL)
    AND g.item_id = ?;

-- name: AdminUpdateForumTopic :exec
UPDATE forumtopic SET title = ?, description = ?, forumcategory_idforumcategory = ?, language_id = sqlc.narg(topic_language_id) WHERE idforumtopic = ?;

-- name: GetAllForumTopicsByCategoryIdForUserWithLastPosterName :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.forumcategory_idforumcategory = sqlc.arg(category_id)
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
      AND ((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;

-- name: ListPrivateTopicParticipantsByTopicIDForUser :many
SELECT u.idusers, u.username
FROM grants g
JOIN users u ON u.idusers = g.user_id
WHERE g.section = 'privateforum'
  AND g.item = 'topic'
  AND g.action = 'view'
  AND g.active = 1
  AND g.user_id IS NOT NULL
  AND g.item_id = sqlc.arg(topic_id)
  AND EXISTS (
      SELECT 1 FROM grants pg
      WHERE pg.section='privateforum'
        AND pg.item='topic'
        AND pg.action='view'
        AND pg.active=1
        AND pg.item_id = g.item_id
        AND pg.user_id = sqlc.arg(viewer_id)
  );

-- name: AdminListPrivateTopicParticipantsByTopicID :many
SELECT u.idusers, u.username
FROM grants g
JOIN users u ON u.idusers = g.user_id
WHERE g.section = 'privateforum'
  AND g.item = 'topic'
  AND g.action = 'view'
  AND g.active = 1
  AND g.user_id IS NOT NULL
  AND g.item_id = ?;

-- name: SystemCopyPrivateTopicGrantsToThread :exec
INSERT INTO grants (
    created_at, user_id, role_id, section, item, rule_type,
    item_id, item_rule, action, extra, active
)
SELECT DISTINCT
    NOW(), topic_grant.user_id, topic_grant.role_id,
    'privateforum_thread', 'thread', 'allow',
    thread_row.idforumthread, NULL, topic_grant.action, NULL, 1
FROM grants topic_grant
JOIN forumtopic topic
    ON topic.idforumtopic = topic_grant.item_id
   AND topic.handler = 'private'
JOIN forumthread thread_row
    ON thread_row.idforumthread = sqlc.arg(thread_id)
   AND thread_row.forumtopic_idforumtopic = topic.idforumtopic
WHERE topic.idforumtopic = sqlc.arg(topic_id)
  AND topic_grant.section = 'privateforum'
  AND topic_grant.item = 'topic'
  AND topic_grant.rule_type = 'allow'
  AND topic_grant.active = 1
  AND topic_grant.action IN ('view', 'reply')
  AND (topic_grant.user_id IS NOT NULL OR topic_grant.role_id IS NOT NULL)
  AND NOT EXISTS (
      SELECT 1
      FROM grants thread_grant
      WHERE thread_grant.section = 'privateforum_thread'
        AND thread_grant.item = 'thread'
        AND thread_grant.rule_type = 'allow'
        AND thread_grant.item_id = thread_row.idforumthread
        AND thread_grant.action = topic_grant.action
        AND thread_grant.active = 1
        AND (thread_grant.user_id <=> topic_grant.user_id)
        AND (thread_grant.role_id <=> topic_grant.role_id)
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
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
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
      AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY t.lastaddition DESC;

-- name: GetForumTopicByIdForUser :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT t.*, lu.username AS LastPosterUsername
FROM forumtopic t
LEFT JOIN users lu ON lu.idusers = t.lastposter
WHERE t.idforumtopic = sqlc.arg(idforumtopic)
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
    WHERE ((t.handler = 'private' AND g.section = 'privateforum') OR (t.handler <> 'private' AND g.section = 'forum'))
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

-- name: AdminCreateForumCategory :execlastid
INSERT INTO forumcategory (forumcategory_idforumcategory, language_id, title, description)
VALUES (sqlc.arg(parent_id), sqlc.narg(category_language_id), sqlc.arg(title), sqlc.arg(description));

INSERT INTO forumtopic (forumcategory_idforumcategory, language_id, title, description, handler) VALUES (?, sqlc.narg(topic_language_id), ?, ?, ?);

-- name: CreateForumTopicForPoster :execlastid
INSERT INTO forumtopic (forumcategory_idforumcategory, language_id, title, description, handler)
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
SELECT t.*
FROM forumtopic t
WHERE t.idforumtopic = sqlc.arg(idforumtopic);

-- name: GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT th.*, lu.username AS lastposterusername, lu.idusers AS lastposterid, fcu.username as firstpostusername, fcu.idusers as firstpostuserid, fc.written as firstpostwritten, fc.text as firstposttext
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic=t.idforumtopic
LEFT JOIN users lu ON lu.idusers = th.lastposter
LEFT JOIN comments fc ON th.firstpost=fc.idcomments
LEFT JOIN users fcu ON fcu.idusers = fc.users_idusers
WHERE th.forumtopic_idforumtopic=sqlc.arg(topic_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE ((t.handler = 'private' AND g.section = 'privateforum') OR (t.handler <> 'private' AND g.section = 'forum'))
      AND (g.item='topic' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND ((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
  AND (
      t.handler IS NULL
      OR t.handler != 'private'
      OR EXISTS (
          SELECT 1 FROM grants thread_grant
          WHERE thread_grant.section = 'privateforum_thread'
            AND thread_grant.item = 'thread'
            AND thread_grant.action = 'view'
            AND thread_grant.active = 1
            AND thread_grant.item_id = th.idforumthread
            AND (thread_grant.user_id = sqlc.arg(viewer_match_id) OR thread_grant.user_id IS NULL)
            AND (thread_grant.role_id IS NULL OR thread_grant.role_id IN (SELECT id FROM role_ids))
      )
  )
ORDER BY th.lastaddition DESC;

-- name: ListPrivateTopicsByUserID :many
SELECT t.idforumtopic,
       t.lastposter,
       t.forumcategory_idforumcategory,
       t.language_id,
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
  AND g.section = 'privateforum'
  AND g.item = 'topic'
  AND g.action = 'see'
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
DELETE FROM forumtopic WHERE idforumtopic = ?;


-- name: GetAllForumThreadsWithTopic :many
SELECT th.*, t.title AS topic_title
FROM forumthread th
LEFT JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
ORDER BY t.idforumtopic, th.lastaddition DESC;

-- name: GetForumCategoryById :one
SELECT * FROM forumcategory
WHERE idforumcategory = sqlc.arg(idforumcategory)
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
WHERE forumcategory_idforumcategory = sqlc.arg(category_id)
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

-- name: AdminCreateForumTopic :execlastid
INSERT INTO forumtopic (forumcategory_idforumcategory, language_id, title, description, handler)
VALUES (sqlc.arg(forumcategory_id), sqlc.narg(language_id), sqlc.arg(title), sqlc.arg(description), sqlc.arg(handler));

-- name: AdminGetTopicGrants :many
SELECT g.section, g.role_id, r.name as role_name, g.user_id, u.username
FROM grants g
LEFT JOIN roles r ON r.id = g.role_id
LEFT JOIN users u ON u.idusers = g.user_id
WHERE (g.item = 'topic')
  AND g.item_id = sqlc.arg(topic_id)
  AND g.active = 1;

-- name: GetPrivateTopicThreadsAndLabelsForUser :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(user_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT th.idforumthread, c.users_idusers AS author_id, cpl.label, cpl.invert
FROM forumthread th
JOIN comments c ON th.firstpost = c.idcomments
LEFT JOIN content_private_labels cpl
    ON cpl.item = 'thread'
    AND cpl.item_id = th.idforumthread
    AND cpl.user_id = sqlc.arg(user_id)
WHERE th.forumtopic_idforumtopic = sqlc.arg(topic_id)
  AND EXISTS (
      SELECT 1 FROM grants thread_grant
      WHERE thread_grant.section = 'privateforum_thread'
        AND thread_grant.item = 'thread'
        AND thread_grant.action = 'view'
        AND thread_grant.active = 1
        AND thread_grant.item_id = th.idforumthread
        AND (thread_grant.user_id = sqlc.narg(viewer_match_id) OR thread_grant.user_id IS NULL)
        AND (thread_grant.role_id IS NULL OR thread_grant.role_id IN (SELECT id FROM role_ids))
  );


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
       th.comments,
       lu.username AS lastposterusername,
       lu.idusers AS lastposterid,
       fcu.username AS firstpostusername,
       fcu.idusers AS firstpostuserid,
       fc.written AS firstpostwritten,
       fc.text AS firstposttext
FROM forumthread th
JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
JOIN comments c ON th.firstpost = c.idcomments
LEFT JOIN users lu ON lu.idusers = th.lastposter
LEFT JOIN comments fc ON th.firstpost = fc.idcomments
LEFT JOIN users fcu ON fcu.idusers = fc.users_idusers
WHERE t.handler = 'private'
  AND (sqlc.arg(topic_id_null) IS NULL OR th.forumtopic_idforumtopic = sqlc.arg(topic_id_val))
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
  AND EXISTS (
      SELECT 1 FROM grants thread_grant
      WHERE thread_grant.section = 'privateforum_thread'
        AND thread_grant.item = 'thread'
        AND thread_grant.action = 'view'
        AND thread_grant.active = 1
        AND thread_grant.item_id = th.idforumthread
        AND (thread_grant.user_id = sqlc.arg(grant_user_id) OR thread_grant.user_id IS NULL)
        AND (thread_grant.role_id IS NULL OR thread_grant.role_id IN (SELECT id FROM role_ids))
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

-- name: CountUnreadPrivateThreadsForUser :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(grantee_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT count(*)
FROM forumthread th
JOIN forumtopic t ON th.forumtopic_idforumtopic = t.idforumtopic
JOIN comments c ON th.firstpost = c.idcomments
WHERE t.handler = 'private'
  AND (sqlc.arg(topic_id_null) IS NULL OR th.forumtopic_idforumtopic = sqlc.arg(topic_id_val))
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
  AND EXISTS (
      SELECT 1 FROM grants thread_grant
      WHERE thread_grant.section = 'privateforum_thread'
        AND thread_grant.item = 'thread'
        AND thread_grant.action = 'view'
        AND thread_grant.active = 1
        AND thread_grant.item_id = th.idforumthread
        AND (thread_grant.user_id = sqlc.arg(grant_user_id) OR thread_grant.user_id IS NULL)
        AND (thread_grant.role_id IS NULL OR thread_grant.role_id IN (SELECT id FROM role_ids))
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
  );

-- name: SetThreadReplyTo :exec
UPDATE forumthread
SET reply_to_comment_id = sqlc.arg(reply_to_comment_id),
    reply_to_thread_id = sqlc.arg(reply_to_thread_id)
WHERE idforumthread = sqlc.arg(idforumthread);

-- name: GetReplyThreadsForThread :many
SELECT t.*,
       c.text as first_post_text,
       t.comments as total_comments,
       t.idforumthread,
       u.username as last_poster_name,
       cu.username as firstpostusername,
       cu.idusers as firstpostuserid,
       c.written as firstpostwritten
FROM forumthread t
LEFT JOIN comments c ON t.firstpost = c.idcomments
LEFT JOIN users u ON t.lastposter = u.idusers
LEFT JOIN users cu ON c.users_idusers = cu.idusers
WHERE t.reply_to_thread_id = sqlc.arg(reply_to_thread_id)
ORDER BY t.lastaddition DESC;

-- name: GetReplyThreadCountsForComments :many
SELECT t.reply_to_comment_id, COUNT(t.idforumthread) as thread_count
FROM forumthread t
WHERE t.reply_to_thread_id = sqlc.arg(reply_to_thread_id)
  AND t.reply_to_comment_id IN (sqlc.slice('comment_ids'))
GROUP BY t.reply_to_comment_id;

-- name: CountReplyThreadsForThread :one
SELECT COUNT(*) FROM forumthread WHERE reply_to_thread_id = sqlc.arg(reply_to_thread_id);

-- name: SystemCopyPrivateThreadGrantsToThread :exec
INSERT INTO grants (
    created_at, user_id, role_id, section, item, rule_type,
    item_id, item_rule, action, extra, active
)
SELECT DISTINCT
    NOW(), src_grant.user_id, src_grant.role_id,
    'privateforum_thread', 'thread', 'allow',
    sqlc.arg(dst_thread_id), NULL, src_grant.action, NULL, 1
FROM grants src_grant
WHERE src_grant.section = 'privateforum_thread'
  AND src_grant.item = 'thread'
  AND src_grant.rule_type = 'allow'
  AND src_grant.active = 1
  AND src_grant.action IN ('view', 'reply')
  AND src_grant.item_id = sqlc.arg(src_thread_id)
  AND (src_grant.user_id IS NOT NULL OR src_grant.role_id IS NOT NULL)
  AND NOT EXISTS (
      SELECT 1
      FROM grants dst_grant
        WHERE dst_grant.section = 'privateforum_thread'
          AND dst_grant.item = 'thread'
          AND dst_grant.rule_type = 'allow'
          AND dst_grant.item_id = sqlc.arg(dst_thread_id)
        AND dst_grant.action = src_grant.action
        AND dst_grant.active = 1
        AND (dst_grant.user_id <=> src_grant.user_id)
        AND (dst_grant.role_id <=> src_grant.role_id)
  );
