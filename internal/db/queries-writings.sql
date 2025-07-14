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

-- name: GetPublicWritingsInCategory :many
SELECT w.*, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers=u.idusers
WHERE w.private = 0 AND w.writing_category_id=?
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
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.*, u.idusers AS WriterId, u.Username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE w.idwriting = sqlc.arg(idwriting) AND (w.private = 0 OR EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
))
ORDER BY w.published DESC
;

-- name: GetWritingsByIdsForUserDescendingByPublishedDate :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT w.*, u.idusers AS WriterId, u.username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE w.idwriting IN (sqlc.slice(writingIds)) AND (w.private = 0 OR EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='article'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = w.idwriting
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  ) OR w.users_idusers = sqlc.arg(viewer_id))
ORDER BY w.published DESC
;

-- name: InsertWritingCategory :exec
INSERT INTO writing_category (writing_category_id, title, description)
VALUES (?, ?, ?);

-- name: UpdateWritingCategory :exec
UPDATE writing_category
SET title = ?, description = ?, writing_category_id = ?
WHERE idwritingCategory = ?;

-- name: GetAllWritingCategoriesForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT wc.*
FROM writing_category wc
WHERE wc.writing_category_id = sqlc.arg(WritingCategoryID)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='writing'
      AND g.item='category'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = wc.idwritingcategory
      AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
    );

-- name: FetchAllCategoriesForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
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
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.users_idusers = ?
ORDER BY w.published DESC;
