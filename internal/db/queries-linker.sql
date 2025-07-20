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

-- name: GetAllLinkerCategoriesForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT
    lc.idlinkerCategory,
    lc.position,
    lc.title,
    lc.sortorder
FROM linker_category lc
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND g.item='category'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = lc.idlinkerCategory
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
ORDER BY lc.position;

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

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE (lc.idlinkerCategory = sqlc.arg(idlinkercategory) OR sqlc.arg(idlinkercategory) = 0)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND g.item='link'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = l.idlinker
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
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

-- name: GetLinkerItemsByUserDescendingForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.comments, lc.title as Category_Title, u.username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE l.users_idusers = sqlc.arg(user_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND g.item='link'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = l.idlinker
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending :one
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
WHERE l.idlinker = ?;

-- name: GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser :one
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
WHERE l.idlinker = sqlc.arg(idlinker)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND g.item='link'
      AND g.action IN ('view','comment','reply')
      AND g.active=1
      AND g.item_id = l.idlinker
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
LIMIT 1;

-- name: GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescending :many
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
WHERE l.idlinker IN (sqlc.slice(linkerIds));

-- name: GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
WHERE l.idlinker IN (sqlc.slice(linkerIds))
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND g.item='link'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = l.idlinker
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

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


-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginated :many
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE (lc.idlinkerCategory = sqlc.arg(idlinkercategory) OR sqlc.arg(idlinkercategory) = 0)
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginated :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE (lc.idlinkerCategory = sqlc.arg(idlinkercategory) OR sqlc.arg(idlinkercategory) = 0)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND g.item='link'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = l.idlinker
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;
