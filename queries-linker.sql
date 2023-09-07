-- name: DeleteCategory :exec
DELETE FROM linkerCategory WHERE idlinkerCategory = ?;

-- name: RenameCategory :exec
UPDATE linkerCategory SET title = ? WHERE idlinkerCategory = ?;

-- name: CreateCategory :exec
INSERT INTO linkerCategory (title) VALUES (?);

-- name: ShowCategories :many
SELECT idlinkerCategory, title FROM linkerCategory;

-- name: AdminCategories :many
SELECT idlinkerCategory, title FROM linkerCategory;

-- name: DeleteQueueItem :exec
DELETE FROM linkerQueue WHERE idlinkerQueue = ?;

-- name: UpdateQueue :exec
UPDATE linkerQueue SET linkerCategory_idlinkerCategory = ?, title = ?, url = ?, description = ? WHERE idlinkerQueue = ?;

-- name: AddToQueue :exec
INSERT INTO linkerQueue (users_idusers, linkerCategory_idlinkerCategory, title, url, description) VALUES (?, ?, ?, ?, ?);

-- name: ShowAdminQueue :many
SELECT l.*, u.username, c.title as category_title, c.idlinkerCategory
FROM linkerQueue l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory c ON l.linkerCategory_idlinkerCategory = c.idlinkerCategory
;
-- name: MoveToLinker :exec
INSERT INTO linker (users_idusers, linkerCategory_idlinkerCategory, language_idlanguage, title, `url`, description)
SELECT l.users_idusers, l.linkerCategory_idlinkerCategory, l.language_idlanguage, l.title, l.url, l.description
FROM linkerQueue l
WHERE l.idlinkerQueue = ?
;

-- name: AddToLinker :exec
INSERT INTO linker (users_idusers, linkerCategory_idlinkerCategory, title, url, description, listed)
VALUES (?, ?, ?, ?, ?, NOW());

-- name: AssignLinkerThisThreadId :exec
UPDATE linker SET forumthread_idforumthread = ? WHERE idlinker = ?;

-- name: ShowLatest :many
SELECT l.*, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_idforumthread = th.idforumthread
WHERE lc.idlinkerCategory = ?
ORDER BY l.listed DESC;

-- name: ShowLink :one
SELECT l.*, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
WHERE l.idlinker = ?;

-- name: ShowLinks :many
SELECT l.*, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
WHERE l.idlinker IN (sqlc.slice(linkerIds));

