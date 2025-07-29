-- name: GetPublicWritings :many
SELECT w.*
FROM writing w
WHERE w.private = 0
ORDER BY w.published DESC
LIMIT ? OFFSET ?
;

-- name: GetPublicWritingsByUser :many
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.private = 0 AND w.users_idusers = ?
ORDER BY w.published DESC
LIMIT ? OFFSET ?;

-- name: GetPublicWritingsByUserForViewer :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.private = 0 AND w.users_idusers = sqlc.arg(author_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
LIMIT ? OFFSET ?;

-- name: GetPublicWritingsInCategory :many
SELECT w.*, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers=u.idusers
WHERE w.private = 0 AND w.writing_category_id=?
ORDER BY w.published DESC
LIMIT ? OFFSET ?
;

-- name: GetPublicWritingsInCategoryForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.*, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers=u.idusers
WHERE w.private = 0 AND w.writing_category_id = sqlc.arg(writing_category_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
LIMIT ? OFFSET ?
;

-- name: UpdateWriting :exec
UPDATE writing
SET title = ?, abstract = ?, writing = ?, private = ?, language_idlanguage = ?
WHERE idwriting = ?;

-- name: InsertWriting :execlastid
INSERT INTO writing (writing_category_id, title, abstract, writing, private, language_idlanguage, published, users_idusers)
VALUES (?, ?, ?, ?, ?, ?, NOW(), ?);

-- name: GetWritingByIdForUserDescendingByPublishedDate :one
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.*, u.idusers AS WriterId, u.Username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE w.idwriting = sqlc.arg(idwriting)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
;

-- name: GetWritingsByIdsForUserDescendingByPublishedDate :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.*, u.idusers AS WriterId, u.username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE w.idwriting IN (sqlc.slice(writing_ids))
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC
;

-- name: InsertWritingCategory :exec
INSERT INTO writing_category (writing_category_id, title, description)
VALUES (?, ?, ?);

-- name: UpdateWritingCategory :exec
UPDATE writing_category
SET title = ?, description = ?, writing_category_id = ?
WHERE idwritingCategory = ?;

-- name: GetWritingCategory :one
SELECT * FROM writing_category WHERE idwritingcategory = ?;

-- name: GetAllWritingCategories :many
SELECT *
FROM writing_category
WHERE writing_category_id = ?;

-- name: FetchAllCategories :many
SELECT wc.*
FROM writing_category wc
;

-- name: FetchCategoriesForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT wc.*
FROM writing_category wc
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='category'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = wc.idwritingcategory
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
);


-- name: AssignWritingThisThreadId :exec
UPDATE writing SET forumthread_id = ? WHERE idwriting = ?;


-- name: GetAllWritingsByUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.users_idusers = sqlc.arg(author_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY w.published DESC;
-- name: ListWritersForViewer :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username, COUNT(w.idwriting) AS count
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE (
    NOT EXISTS (SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id))
    OR w.language_idlanguage IN (
        SELECT ul.language_idlanguage FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
)
AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'writing'
      AND g.item = 'article'
      AND g.action = 'see'
      AND g.active = 1
      AND g.item_id = w.idwriting
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?;

-- name: SearchWritersForViewer :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT u.username, COUNT(w.idwriting) AS count
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE (LOWER(u.username) LIKE LOWER(sqlc.arg(query)) OR LOWER((SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1)) LIKE LOWER(sqlc.arg(query)))
  AND (
    NOT EXISTS (SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id))
    OR w.language_idlanguage IN (
        SELECT ul.language_idlanguage FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
    )
  )
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section = 'writing'
      AND g.item = 'article'
      AND g.action = 'see'
      AND g.active = 1
      AND g.item_id = w.idwriting
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
GROUP BY u.idusers
ORDER BY u.username
LIMIT ? OFFSET ?;

-- name: SetWritingLastIndex :exec
UPDATE writing SET last_index = NOW() WHERE idwriting = ?;


-- name: GetAllWritingsForIndex :many
SELECT idwriting, title, abstract, writing FROM writing WHERE deleted_at IS NULL;

