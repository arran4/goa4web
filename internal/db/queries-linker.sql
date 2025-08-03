-- AdminDeleteLinkerCategory removes a linker category.
-- name: AdminDeleteLinkerCategory :exec
DELETE FROM linker_category
WHERE idlinkerCategory = ?;

-- name: RenameLinkerCategory :exec
UPDATE linker_category SET title = ?, position = ?
WHERE idlinkerCategory = ?
  AND EXISTS (
    SELECT 1 FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.users_idusers = sqlc.arg(admin_id) AND r.is_admin = 1
  );

-- name: AdminCreateLinkerCategory :exec
INSERT INTO linker_category (title, position) VALUES (sqlc.arg(title), sqlc.arg(position));

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
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
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
      AND (g.item='category' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = lc.idlinkerCategory OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
  AND EXISTS (
    SELECT 1 FROM linker l
    WHERE l.linker_category_id = lc.idlinkerCategory
      AND l.listed IS NOT NULL
      AND l.deleted_at IS NULL
  )
ORDER BY lc.position;

-- name: GetLinkerCategoryLinkCounts :many
SELECT c.idlinkerCategory, c.title, c.position, COUNT(l.idlinker) as LinkCount
FROM linker_category c
LEFT JOIN linker l ON c.idlinkerCategory = l.linker_category_id AND l.listed IS NOT NULL AND l.deleted_at IS NULL
GROUP BY c.idlinkerCategory
ORDER BY c.position
;



-- name: AdminDeleteLinkerQueuedItem :exec
DELETE FROM linker_queue
WHERE idlinkerQueue = ?;

-- name: UpdateLinkerQueuedItem :exec
UPDATE linker_queue SET linker_category_id = ?, title = ?, url = ?, description = ?
WHERE idlinkerQueue = ?
  AND EXISTS (
    SELECT 1 FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.users_idusers = sqlc.arg(admin_id) AND r.is_admin = 1
  );

-- name: CreateLinkerQueuedItemForWriter :exec
INSERT INTO linker_queue (users_idusers, linker_category_id, title, url, description)
SELECT sqlc.arg(writer_id), sqlc.arg(linker_category_id), sqlc.arg(title), sqlc.arg(url), sqlc.arg(description)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND (g.item='link' OR g.item IS NULL)
      AND g.action='post'
      AND g.active=1
      AND (g.item_id = sqlc.arg(grant_category_id) OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(writer_id)
      ))
);

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
  AND EXISTS (
    SELECT 1 FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.users_idusers = sqlc.arg(admin_id) AND r.is_admin = 1
  );

-- name: AdminCreateLinkerItem :exec
INSERT INTO linker (users_idusers, linker_category_id, title, url, description, listed)
VALUES (sqlc.arg(users_idusers), sqlc.arg(linker_category_id), sqlc.arg(title), sqlc.arg(url), sqlc.arg(description), NOW());

-- name: UpdateLinkerItem :exec
UPDATE linker SET title = ?, url = ?, description = ?, linker_category_id = ?, language_idlanguage = ?
WHERE idlinker = ?
  AND EXISTS (
    SELECT 1 FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.users_idusers = sqlc.arg(admin_id) AND r.is_admin = 1
  );

-- name: SystemAssignLinkerThreadID :exec
UPDATE linker SET forumthread_id = ? WHERE idlinker = ?;

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending :many
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE (lc.idlinkerCategory = sqlc.arg(idlinkercategory) OR sqlc.arg(idlinkercategory) = 0)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
ORDER BY l.listed DESC;

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE (lc.idlinkerCategory = sqlc.arg(idlinkercategory) OR sqlc.arg(idlinkercategory) = 0)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND (g.item='link' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = l.idlinker OR g.item_id IS NULL)
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

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow :many
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE (lc.idlinkerCategory = sqlc.arg(idlinkercategory) OR sqlc.arg(idlinkercategory) = 0)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND (g.item='link' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = l.idlinker OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetLinkerItemsByUserDescendingForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, th.comments, lc.title as Category_Title, u.username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.users_idusers = u.idusers
LEFT JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
LEFT JOIN forumthread th ON l.forumthread_id = th.idforumthread
WHERE l.users_idusers = sqlc.arg(user_id)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='linker'
      AND (g.item='link' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = l.idlinker OR g.item_id IS NULL)
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
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
WHERE l.idlinker = sqlc.arg(idlinker)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
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
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
)
SELECT l.idlinker, l.language_idlanguage, l.users_idusers, l.linker_category_id, l.forumthread_id, l.title, l.url, l.description, l.listed, u.username, lc.title
FROM linker l
JOIN users u ON l.users_idusers = u.idusers
JOIN linker_category lc ON l.linker_category_id = lc.idlinkerCategory
WHERE l.idlinker IN (sqlc.slice(linkerIds))
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
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
LEFT JOIN linker l ON l.linker_category_id = c.idlinkerCategory AND l.listed IS NOT NULL AND l.deleted_at IS NULL
GROUP BY c.idlinkerCategory
ORDER BY c.sortorder;

-- name: UpdateLinkerCategorySortOrder :exec
UPDATE linker_category SET sortorder = ?
WHERE idlinkerCategory = ?
  AND EXISTS (
    SELECT 1 FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.users_idusers = sqlc.arg(admin_id) AND r.is_admin = 1
  );

-- name: AdminCountLinksByCategory :one
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
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
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
      AND (g.item='link' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = l.idlinker OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetAllLinkersForIndex :many
SELECT idlinker, title, description FROM linker WHERE deleted_at IS NULL AND listed IS NOT NULL;

-- name: GetLinkerCategoryById :one
SELECT * FROM linker_category WHERE idlinkerCategory = ?;
