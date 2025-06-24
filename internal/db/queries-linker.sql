-- name: DeleteLinkerCategory :exec
DELETE FROM linkerCategory WHERE idlinkerCategory = ?;

-- name: RenameLinkerCategory :exec
UPDATE linkerCategory SET title = ?, position = ? WHERE idlinkerCategory = ?;

-- name: CreateLinkerCategory :exec
INSERT INTO linkerCategory (title, position) VALUES (?, ?);

-- name: GetAllLinkerCategories :many
SELECT
    lc.idlinkerCategory,
    lc.position,
    lc.title,
    lc.sortorder
FROM linkerCategory lc
ORDER BY lc.position
;

-- name: GetLinkerCategoryLinkCounts :many
SELECT c.idlinkerCategory, c.title, c.position, COUNT(l.idlinker) as LinkCount
FROM linkerCategory c
LEFT JOIN linker l ON c.idlinkerCategory = l.linkerCategory_idlinkerCategory
GROUP BY c.idlinkerCategory
ORDER BY c.position
;

-- name: GetAllLinkerCategoriesWithSortOrder :many
SELECT
    idlinkerCategory,
    position,
    title,
    sortorder
FROM linkerCategory
ORDER BY sortorder;


-- name: DeleteLinkerQueuedItem :exec
DELETE FROM linkerQueue WHERE idlinkerQueue = ?;

-- name: UpdateLinkerQueuedItem :exec
UPDATE linkerQueue SET linkerCategory_idlinkerCategory = ?, title = ?, url = ?, description = ? WHERE idlinkerQueue = ?;

-- name: CreateLinkerQueuedItem :exec
INSERT INTO linkerQueue (users_idusers, linkerCategory_idlinkerCategory, title, url, description) VALUES (?, ?, ?, ?, ?);

-- name: GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails :many
SELECT l.*, u.username, c.title as category_title, c.idlinkerCategory
FROM linkerQueue l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory c ON l.linkerCategory_idlinkerCategory = c.idlinkerCategory
;
-- name: SelectInsertLInkerQueuedItemIntoLinkerByLinkerQueueId :execlastid
INSERT INTO linker (users_idusers, linkerCategory_idlinkerCategory, language_idlanguage, title, `url`, description)
SELECT l.users_idusers, l.linkerCategory_idlinkerCategory, l.language_idlanguage, l.title, l.url, l.description
FROM linkerQueue l
WHERE l.idlinkerQueue = ?
;

-- name: CreateLinkerItem :exec
INSERT INTO linker (users_idusers, linkerCategory_idlinkerCategory, title, url, description, listed)
VALUES (?, ?, ?, ?, ?, NOW());

-- name: AssignLinkerThisThreadId :exec
UPDATE linker SET forumthread_idforumthread = ? WHERE idlinker = ?;

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending :many
SELECT l.*, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_idforumthread = th.idforumthread
WHERE (lc.idlinkerCategory = sqlc.arg(idlinkercategory) OR sqlc.arg(idlinkercategory) = 0)
ORDER BY l.listed DESC;

-- name: GetLinkerItemsByUserDescending :many
SELECT l.*, th.comments, lc.title as Category_Title, u.username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_idforumthread = th.idforumthread
WHERE l.users_idusers = ?
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending :one
SELECT l.*, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
WHERE l.idlinker = ?;

-- name: GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescending :many
SELECT l.*, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linkerCategory lc ON l.linkerCategory_idlinkerCategory = lc.idlinkerCategory
WHERE l.idlinker IN (sqlc.slice(linkerIds));

-- name: GetLinkerCategoriesWithCount :many
SELECT c.idlinkerCategory, c.title, c.sortorder, COUNT(l.idlinker) AS linkcount
FROM linkerCategory c
LEFT JOIN linker l ON l.linkerCategory_idlinkerCategory = c.idlinkerCategory
GROUP BY c.idlinkerCategory
ORDER BY c.sortorder;

-- name: UpdateLinkerCategorySortOrder :exec
UPDATE linkerCategory SET sortorder = ? WHERE idlinkerCategory = ?;

-- name: CountLinksByCategory :one
SELECT COUNT(*) FROM linker WHERE linkerCategory_idlinkerCategory = ?;

