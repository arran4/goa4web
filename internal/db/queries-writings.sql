-- name: GetPublicWritings :many
SELECT w.*
FROM writing w
WHERE w.private = 0
ORDER BY w.published DESC
LIMIT ? OFFSET ?
;

-- name: GetPublicWritingsByUser :many
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_idforumthread=w.forumthread_idforumthread AND w.forumthread_idforumthread != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.private = 0 AND w.users_idusers = ?
ORDER BY w.published DESC
LIMIT ? OFFSET ?;

-- name: GetPublicWritingsInCategory :many
SELECT w.*, u.Username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_idforumthread=w.forumthread_idforumthread AND w.forumthread_idforumthread != 0) as Comments
FROM writing w
LEFT JOIN users u ON w.Users_Idusers=u.idusers
WHERE w.private = 0 AND w.writingCategory_idwritingCategory=?
ORDER BY w.published DESC
LIMIT ? OFFSET ?
;

-- name: UpdateWriting :exec
UPDATE writing
SET title = ?, abstract = ?, writing = ?, private = ?, language_idlanguage = ?
WHERE idwriting = ?;

-- name: InsertWriting :execlastid
INSERT INTO writing (writingCategory_idwritingCategory, title, abstract, writing, private, language_idlanguage, published, users_idusers)
VALUES (?, ?, ?, ?, ?, ?, NOW(), ?);

-- name: GetWritingByIdForUserDescendingByPublishedDate :one
SELECT w.*, u.idusers AS WriterId, u.Username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
LEFT JOIN writingApprovedUsers wau ON w.idwriting = wau.writing_idwriting AND wau.users_idusers = sqlc.arg(UserId)
WHERE w.idwriting = ? AND (w.private = 0 OR wau.readdoc = 1 OR w.users_idusers = sqlc.arg(UserId))
ORDER BY w.published DESC
;

-- name: GetWritingsByIdsForUserDescendingByPublishedDate :many
SELECT w.*, u.idusers AS WriterId, u.username AS WriterUsername
FROM writing w
JOIN users u ON w.users_idusers = u.idusers
LEFT JOIN writingApprovedUsers wau ON w.idwriting = wau.writing_idwriting AND wau.users_idusers = sqlc.arg(userId)
WHERE w.idwriting IN (sqlc.slice(writingIds)) AND (w.private = 0 OR wau.readdoc = 1 OR w.users_idusers = sqlc.arg(userId))
ORDER BY w.published DESC
;

-- name: InsertWritingCategory :exec
INSERT INTO writingCategory (writingCategory_idwritingCategory, title, description)
VALUES (?, ?, ?);

-- name: UpdateWritingCategory :exec
UPDATE writingCategory
SET title = ?, description = ?, writingCategory_idwritingCategory = ?
WHERE idwritingCategory = ?;

-- name: GetAllWritingCategories :many
SELECT *
FROM writingCategory
WHERE writingCategory_idwritingCategory = ?;

-- name: FetchAllCategories :many
SELECT wc.*
FROM writingCategory wc
;

-- name: DeleteWritingApproval :exec
DELETE FROM writingApprovedUsers
WHERE writing_idwriting = ? AND users_idusers = ?;

-- name: CreateWritingApproval :exec
INSERT INTO writingApprovedUsers (writing_idwriting, users_idusers, readdoc, editdoc)
VALUES (?, ?, ?, ?);

-- name: UpdateWritingApproval :exec
UPDATE writingApprovedUsers
SET readdoc = ?, editdoc = ?
WHERE writing_idwriting = ? AND users_idusers = ?;

-- name: GetAllWritingApprovals :many
SELECT idusers, u.username, wau.*
FROM writingApprovedUsers wau
LEFT JOIN users u ON idusers = wau.users_idusers
;

-- name: AssignWritingThisThreadId :exec
UPDATE writing SET forumthread_idforumthread = ? WHERE idwriting = ?;


-- name: GetAllWritingsByUser :many
SELECT w.*, u.username,
    (SELECT COUNT(*) FROM comments c WHERE c.forumthread_idforumthread=w.forumthread_idforumthread AND w.forumthread_idforumthread != 0) AS Comments
FROM writing w
LEFT JOIN users u ON w.users_idusers = u.idusers
WHERE w.users_idusers = ?
ORDER BY w.published DESC;
