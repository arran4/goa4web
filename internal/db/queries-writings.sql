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
SELECT w.*, u.idusers AS WriterId, u.Username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE (
    w.private = 0 OR w.users_idusers = sqlc.arg(user_id) OR EXISTS (
        SELECT 1 FROM grants g
        WHERE g.user_id = sqlc.arg(viewer_id)
          AND g.section = 'writing'
          AND g.item_id = w.idwriting
          AND g.action = 'view'
          AND g.active = 1
    )
) AND w.idwriting = sqlc.arg(idwriting)
ORDER BY w.published DESC
;

-- name: GetWritingsByIdsForUserDescendingByPublishedDate :many
SELECT w.*, u.idusers AS WriterId, u.username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
WHERE (
    w.private = 0 OR w.users_idusers = sqlc.arg(user_id) OR EXISTS (
        SELECT 1 FROM grants g
        WHERE g.user_id = sqlc.arg(viewer_id)
          AND g.section = 'writing'
          AND g.item_id = w.idwriting
          AND g.action = 'view'
          AND g.active = 1
    )
) AND w.idwriting IN (sqlc.slice(writingIds))
ORDER BY w.published DESC
;

-- name: InsertWritingCategory :exec
INSERT INTO writing_category (writing_category_id, title, description)
VALUES (?, ?, ?);

-- name: UpdateWritingCategory :exec
UPDATE writing_category
SET title = ?, description = ?, writing_category_id = ?
WHERE idwritingCategory = ?;

-- name: GetAllWritingCategories :many
SELECT *
FROM writing_category
WHERE writing_category_id = ?;

-- name: FetchAllCategories :many
SELECT wc.*
FROM writing_category wc
;


-- name: AssignWritingThisThreadId :exec
UPDATE writing SET forumthread_id = ? WHERE idwriting = ?;


-- name: GetAllWritingsByUser :many
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_id=w.forumthread_id AND w.forumthread_id != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.users_idusers = sqlc.arg(author_id)
  AND (
      w.private = 0 OR w.users_idusers = sqlc.arg(viewer_match_id) OR EXISTS (
          SELECT 1 FROM grants g
          WHERE g.user_id = sqlc.arg(viewer_id)
            AND g.section = 'writing'
            AND g.item_id = w.idwriting
            AND g.action = 'view'
            AND g.active = 1
      )
  )
ORDER BY w.published DESC;
