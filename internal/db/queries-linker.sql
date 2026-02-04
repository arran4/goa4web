-- AdminDeleteLinkerCategory removes a linker category.
-- name: AdminDeleteLinkerCategory :exec
DELETE FROM linker_category
WHERE id = ?;

-- name: AdminRenameLinkerCategory :exec
UPDATE linker_category SET title = ?, position = ?
WHERE id = ?;

-- name: AdminCreateLinkerCategory :exec
INSERT INTO linker_category (title, position) VALUES (sqlc.arg(title), sqlc.arg(position));

-- name: GetAllLinkerCategories :many
SELECT
    lc.id,
    lc.position,
    lc.title,
    lc.sortorder
FROM linker_category lc
ORDER BY lc.position
;

-- name: GetAllLinkerCategoriesForUser :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
),
grants_for_viewer AS (
    SELECT g.section, g.item, g.action, g.item_id
    FROM grants g
    WHERE g.active = 1
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
SELECT
    lc.id,
    lc.position,
    lc.title,
    lc.sortorder
FROM linker_category lc
WHERE EXISTS (
    SELECT 1
    FROM grants_for_viewer g
    WHERE g.section = 'linker'
      AND (g.item = 'category' OR g.item IS NULL)
      AND g.action = 'see'
      AND (g.item_id = lc.id OR g.item_id IS NULL)
)
  AND EXISTS (
    SELECT 1 FROM linker l
    WHERE l.category_id = lc.id
      AND l.listed IS NOT NULL
      AND l.deleted_at IS NULL
  )
ORDER BY lc.position;

-- name: GetLinkerCategoryLinkCounts :many
SELECT c.id, c.title, c.position, COUNT(l.id) as LinkCount
FROM linker_category c
LEFT JOIN linker l ON c.id = l.category_id AND l.listed IS NOT NULL AND l.deleted_at IS NULL
GROUP BY c.id
ORDER BY c.position
;



-- name: AdminDeleteLinkerQueuedItem :exec
DELETE FROM linker_queue
WHERE id = ?;

-- name: AdminUpdateLinkerQueuedItem :exec
UPDATE linker_queue SET category_id = ?, title = ?, url = ?, description = ?
WHERE id = ?;

-- name: CreateLinkerQueuedItemForWriter :exec
INSERT INTO linker_queue (submitter_id, category_id, title, url, description, timezone)
SELECT sqlc.arg(writer_id), sqlc.arg(category_id), sqlc.arg(title), sqlc.arg(url), sqlc.arg(description), sqlc.arg(timezone)
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
SELECT l.*, u.username, c.title as category_title, c.id AS category_id
FROM linker_queue l
JOIN users u ON l.submitter_id = u.idusers
JOIN linker_category c ON l.category_id = c.id
;
-- name: AdminInsertQueuedLinkFromQueue :execlastid
INSERT INTO linker (author_id, category_id, language_id, title, `url`, description, timezone)
SELECT l.submitter_id, l.category_id, l.language_id, l.title, l.url, l.description, l.timezone
FROM linker_queue l
WHERE l.id = ?;

-- name: AdminCreateLinkerItem :exec
INSERT INTO linker (author_id, category_id, title, url, description, listed, timezone)
VALUES (sqlc.arg(author_id), sqlc.arg(category_id), sqlc.arg(title), sqlc.arg(url), sqlc.arg(description), sqlc.arg(listed), sqlc.arg(timezone));

-- name: AdminUpdateLinkerItem :exec
UPDATE linker SET title = ?, url = ?, description = ?, category_id = ?, language_id = ?
WHERE id = ?;

-- name: SystemAssignLinkerThreadID :exec
UPDATE linker SET thread_id = ? WHERE id = ?;

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending :many
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.author_id = u.idusers
LEFT JOIN linker_category lc ON l.category_id = lc.id
LEFT JOIN forumthread th ON l.thread_id = th.idforumthread
WHERE (lc.id = sqlc.arg(category_id) OR sqlc.arg(category_id) = 0)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
ORDER BY l.listed DESC;

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUser :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
),
grants_for_viewer AS (
    SELECT g.section, g.item, g.action, g.item_id
    FROM grants g
    WHERE g.active = 1
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.author_id = u.idusers
LEFT JOIN linker_category lc ON l.category_id = lc.id
LEFT JOIN forumthread th ON l.thread_id = th.idforumthread
WHERE (lc.id = sqlc.arg(category_id) OR sqlc.arg(category_id) = 0)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
  AND EXISTS (
    SELECT 1
    FROM grants_for_viewer g
    WHERE g.section = 'linker'
      AND (g.item = 'link' OR g.item IS NULL)
      AND g.action = 'see'
      AND (g.item_id = l.id OR g.item_id IS NULL)
  )
ORDER BY l.listed DESC;

-- name: GetLinkerItemsByUserDescending :many
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, th.comments, lc.title as Category_Title, u.username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.author_id = u.idusers
LEFT JOIN linker_category lc ON l.category_id = lc.id
LEFT JOIN forumthread th ON l.thread_id = th.idforumthread
WHERE l.author_id = ?
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
),
grants_for_viewer AS (
    SELECT g.section, g.item, g.action, g.item_id
    FROM grants g
    WHERE g.active = 1
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.author_id = u.idusers
LEFT JOIN linker_category lc ON l.category_id = lc.id
LEFT JOIN forumthread th ON l.thread_id = th.idforumthread
WHERE (lc.id = sqlc.arg(category_id) OR sqlc.arg(category_id) = 0)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
  AND EXISTS (
    SELECT 1
    FROM grants_for_viewer g
    WHERE g.section = 'linker'
      AND (g.item = 'link' OR g.item IS NULL)
      AND g.action = 'see'
      AND (g.item_id = l.id OR g.item_id IS NULL)
  )
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetLinkerItemsByUserDescendingForUser :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
),
grants_for_viewer AS (
    SELECT g.section, g.item, g.action, g.item_id
    FROM grants g
    WHERE g.active = 1
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, th.comments, lc.title as Category_Title, u.username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.author_id = u.idusers
LEFT JOIN linker_category lc ON l.category_id = lc.id
LEFT JOIN forumthread th ON l.thread_id = th.idforumthread
WHERE l.author_id = sqlc.arg(user_id)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
  AND EXISTS (
    SELECT 1
    FROM grants_for_viewer g
    WHERE g.section = 'linker'
      AND (g.item = 'link' OR g.item IS NULL)
      AND g.action = 'see'
      AND (g.item_id = l.id OR g.item_id IS NULL)
  )
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending :one
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, u.username, lc.title
FROM linker l
JOIN users u ON l.author_id = u.idusers
JOIN linker_category lc ON l.category_id = lc.id
WHERE l.id = ?;

-- name: GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
),
grants_for_viewer AS (
    SELECT g.section, g.item, g.action, g.item_id
    FROM grants g
    WHERE g.active = 1
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, u.username, lc.title
FROM linker l
JOIN users u ON l.author_id = u.idusers
JOIN linker_category lc ON l.category_id = lc.id
WHERE l.id = sqlc.arg(id)
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
  AND EXISTS (
    SELECT 1
    FROM grants_for_viewer g
    WHERE g.section = 'linker'
      AND (g.item = 'link' OR g.item IS NULL)
      AND g.action IN ('view')
      AND (g.item_id = l.id OR g.item_id IS NULL)
  )
LIMIT 1;

-- name: GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescending :many
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, u.username, lc.title
FROM linker l
JOIN users u ON l.author_id = u.idusers
JOIN linker_category lc ON l.category_id = lc.id
WHERE l.id IN (sqlc.slice(linkerIds));

-- name: GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingForUser :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
),
grants_for_viewer AS (
    SELECT g.section, g.item, g.action, g.item_id
    FROM grants g
    WHERE g.active = 1
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, u.username, lc.title
FROM linker l
JOIN users u ON l.author_id = u.idusers
JOIN linker_category lc ON l.category_id = lc.id
WHERE l.id IN (sqlc.slice(linkerIds))
  AND l.listed IS NOT NULL
  AND l.deleted_at IS NULL
  AND EXISTS (
    SELECT 1
    FROM grants_for_viewer g
    WHERE g.section = 'linker'
      AND (g.item = 'link' OR g.item IS NULL)
      AND g.action = 'view'
      AND (g.item_id = l.id OR g.item_id IS NULL)
  );

-- name: GetLinkerCategoriesWithCount :many
SELECT c.id, c.title, c.sortorder, COUNT(l.id) AS linkcount
FROM linker_category c
LEFT JOIN linker l ON l.category_id = c.id AND l.listed IS NOT NULL AND l.deleted_at IS NULL
GROUP BY c.id
ORDER BY c.sortorder;

-- name: AdminUpdateLinkerCategorySortOrder :exec
UPDATE linker_category SET sortorder = ?
WHERE id = ?;

-- name: AdminCountLinksByCategory :one
SELECT COUNT(*) FROM linker WHERE category_id = ?;


-- name: SystemSetLinkerLastIndex :exec
UPDATE linker SET last_index = NOW() WHERE id = ?;


-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingPaginated :many
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.author_id = u.idusers
LEFT JOIN linker_category lc ON l.category_id = lc.id
LEFT JOIN forumthread th ON l.thread_id = th.idforumthread
WHERE (lc.id = sqlc.arg(category_id) OR sqlc.arg(category_id) = 0)
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginated :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
),
grants_for_viewer AS (
    SELECT g.section, g.item, g.action, g.item_id
    FROM grants g
    WHERE g.active = 1
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
)
SELECT l.id, l.language_id, l.author_id, l.category_id, l.thread_id, l.title, l.url, l.description, l.listed, l.timezone, th.Comments, lc.title as Category_Title, u.Username as PosterUsername
FROM linker l
LEFT JOIN users u ON l.author_id = u.idusers
LEFT JOIN linker_category lc ON l.category_id = lc.id
LEFT JOIN forumthread th ON l.thread_id = th.idforumthread
WHERE (lc.id = sqlc.arg(category_id) OR sqlc.arg(category_id) = 0)
  AND EXISTS (
    SELECT 1
    FROM grants_for_viewer g
    WHERE g.section = 'linker'
      AND (g.item = 'link' OR g.item IS NULL)
      AND g.action = 'see'
      AND (g.item_id = l.id OR g.item_id IS NULL)
  )
ORDER BY l.listed DESC
LIMIT ? OFFSET ?;

-- name: GetAllLinkersForIndex :many
SELECT id, title, description FROM linker WHERE deleted_at IS NULL AND listed IS NOT NULL;

-- name: GetLinkerCategoryById :one
SELECT * FROM linker_category WHERE id = ?;

-- name: ListLinkerCategoryPath :many
WITH RECURSIVE category_path AS (
    SELECT lc.id, NULL AS parent_id, lc.title, 0 AS depth
    FROM linker_category lc
    WHERE lc.id = sqlc.arg(category_id)
)
SELECT category_path.id, category_path.title
FROM category_path;
