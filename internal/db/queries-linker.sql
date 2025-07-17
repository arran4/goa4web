-- name: DeleteLinkerCategory :exec
DELETE FROM linker_category WHERE idlinkerCategory = ?;

-- name: RenameLinkerCategory :exec
UPDATE linker_category SET title = ?, position = ? WHERE idlinkerCategory = ?;

-- name: CreateLinkerCategory :exec
INSERT INTO linker_category (title, position) VALUES (?, ?);

-- name: GetAllLinkerCategories :many
SELECT
    lc.idlinkerCategory,
    lc.position,
    lc.title,
    lc.sortorder
FROM linker_category lc
ORDER BY lc.position
;

-- name: GetLinkerCategoryLinkCounts :many
SELECT c.idlinkerCategory, c.title, c.position, COUNT(l.idlinker) as LinkCount
FROM linker_category c
LEFT JOIN linker l ON c.idlinkerCategory = l.linker_category_id
GROUP BY c.idlinkerCategory
ORDER BY c.position
;

-- name: GetAllLinkerCategoriesWithSortOrder :many
SELECT
    idlinkerCategory,
    position,
    title,
    sortorder
FROM linker_category
ORDER BY sortorder;


-- name: DeleteLinkerQueuedItem :exec
DELETE FROM linker_queue WHERE idlinkerQueue = ?;

-- name: UpdateLinkerQueuedItem :exec
UPDATE linker_queue SET linker_category_id = ?, title = ?, url = ?, description = ? WHERE idlinkerQueue = ?;

-- name: CreateLinkerQueuedItem :exec
INSERT INTO linker_queue (users_idusers, linker_category_id, title, url, description) VALUES (?, ?, ?, ?, ?);

-- name: GetAllLinkerQueuedItemsWithUserAndLinkerCategoryDetails :many
SELECT l.*, u.username, c.title as category_title, c.idlinkerCategory
FROM linker_queue l
JOIN users u ON l.users_idusers = u.idusers
JOIN linker_category c ON l.linker_category_id = c.idlinkerCategory
;
-- name: SelectInsertLInkerQueuedItemIntoLinkerByLinkerQueueId :execlastid
INSERT INTO linker (users_idusers, linker_category_id, language_idlanguage, title, `url`, description)
SELECT l.users_idusers, l.linker_category_id, l.language_idlanguage, l.title, l.url, l.description
FROM linker_queue l
WHERE l.idlinkerQueue = ?
;

-- name: CreateLinkerItem :exec
INSERT INTO linker (users_idusers, linker_category_id, title, url, description, listed)
VALUES (?, ?, ?, ?, ?, NOW());

-- name: UpdateLinkerItem :exec
UPDATE linker SET title = ?, url = ?, description = ?, linker_category_id = ?, language_idlanguage = ?
WHERE idlinker = ?;

-- name: AssignLinkerThisThreadId :exec
UPDATE linker SET forumthread_id = ? WHERE idlinker = ?;

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending :many
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE (lc.idlinkerCategory = sqlc.arg(idlinkercategory) OR sqlc.arg(idlinkercategory) = 0)
ORDER BY l.listed DESC;

-- name: GetLinkerItemsByUserDescending :many
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.comments, lc.title as Category_Title, u.username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE l.users_idusers = ?
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending :one
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
WHERE l.idlinker = ?;

-- name: GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescending :many
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
WHERE l.idlinker IN (sqlc.slice(linkerIds));

-- name: GetLinkerCategoriesWithCount :many
SELECT c.idlinkerCategory, c.title, c.sortorder, COUNT(l.idlinker) AS linkcount
FROM linker_category c
LEFT JOIN linker l ON l.linker_category_id = c.idlinkerCategory
GROUP BY c.idlinkerCategory
ORDER BY c.sortorder;

-- name: UpdateLinkerCategorySortOrder :exec
UPDATE linker_category SET sortorder = ? WHERE idlinkerCategory = ?;

-- name: CountLinksByCategory :one
SELECT COUNT(*) FROM linker WHERE linker_category_id = ?;


-- name: SetLinkerLastIndex :exec
UPDATE linker SET last_index = NOW() WHERE idlinker = ?;

