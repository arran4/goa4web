-- name: AdminCreateImageBoard :exec
INSERT INTO imageboard (imageboard_idimageboard, title, description, approval_required) VALUES (?, ?, ?, ?);

-- name: AdminUpdateImageBoard :exec
UPDATE imageboard SET title = ?, description = ?, imageboard_idimageboard = ?, approval_required = ? WHERE idimageboard = ?;

-- name: SystemListBoardsByParentID :many
SELECT b.idimageboard, b.imageboard_idimageboard, b.title, b.description, b.approval_required, b.deleted_at
FROM imageboard b
WHERE (b.imageboard_idimageboard = sqlc.narg(parent_id) OR (b.imageboard_idimageboard IS NULL AND sqlc.narg(parent_id) IS NULL))
  AND b.deleted_at IS NULL
LIMIT ? OFFSET ?;


-- name: CreateImagePostForPoster :execlastid
INSERT INTO imagepost (
    imageboard_idimageboard,
    thumbnail,
    fullimage,
    users_idusers,
    description,
    posted,
    timezone,
    approved,
    file_size
)
SELECT sqlc.arg(imageboard_id), sqlc.arg(thumbnail), sqlc.arg(fullimage), sqlc.arg(poster_id), sqlc.arg(description), sqlc.arg(posted), sqlc.arg(timezone), sqlc.arg(approved), sqlc.arg(file_size)
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND (g.item='board' OR g.item IS NULL)
      AND g.action='post'
      AND g.active=1
      AND (g.item_id = sqlc.arg(grant_board_id) OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(poster_id)
      ))
);

-- name: SystemAssignImagePostThreadID :exec
UPDATE imagepost SET forumthread_id = ? WHERE idimagepost = ?;

-- name: GetImagePostsByUserDescending :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.users_idusers = ? AND i.approved = 1 AND i.deleted_at IS NULL
ORDER BY i.posted DESC
LIMIT ? OFFSET ?;

-- name: GetImagePostsByUserDescendingAll :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.users_idusers = ? AND i.deleted_at IS NULL
ORDER BY i.posted DESC
LIMIT ? OFFSET ?;

-- name: AdminListBoards :many
SELECT b.idimageboard, b.imageboard_idimageboard, b.title, b.description, b.approval_required, b.deleted_at
FROM imageboard b
LIMIT ? OFFSET ?;

-- name: GetImageBoardById :one
SELECT * FROM imageboard WHERE idimageboard = ? AND deleted_at IS NULL;

-- name: AdminDeleteImageBoard :exec
UPDATE imageboard SET deleted_at = NOW() WHERE idimageboard = ?;

-- name: AdminApproveImagePost :exec
UPDATE imagepost SET approved = 1 WHERE idimagepost = ?;


-- name: AdminGetImagePost :one
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.idimagepost = ?
LIMIT 1;

-- name: AdminUpdateImagePost :exec
UPDATE imagepost
SET imageboard_idimageboard = ?, description = ?, approved = ?
WHERE idimagepost = ?;

-- name: AdminDeleteImagePost :exec
UPDATE imagepost SET deleted_at = NOW() WHERE idimagepost = ?;


-- name: ListBoardsByParentIDForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT b.idimageboard, b.imageboard_idimageboard, b.title, b.description, b.approval_required, b.deleted_at
FROM imageboard b
WHERE (b.imageboard_idimageboard = sqlc.narg(parent_id) OR (b.imageboard_idimageboard IS NULL AND sqlc.narg(parent_id) IS NULL))
  AND b.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND (g.item='board' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = b.idimageboard OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(lister_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
LIMIT ? OFFSET ?;

-- name: ListBoardsForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT b.idimageboard, b.imageboard_idimageboard, b.title, b.description, b.approval_required, b.deleted_at
FROM imageboard b
WHERE b.deleted_at IS NULL AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND (g.item='board' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = b.idimageboard OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(lister_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
LIMIT ? OFFSET ?;

-- name: ListImagePostsByPosterForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.users_idusers = sqlc.arg(poster_id)
  AND i.approved = 1
  AND i.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND (g.item='board' OR g.item IS NULL)
      AND g.action='see'
      AND g.active=1
      AND (g.item_id = i.imageboard_idimageboard OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(lister_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY i.posted DESC
LIMIT ? OFFSET ?;

-- name: ListImagePostsByBoardForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.imageboard_idimageboard = sqlc.arg(board_id)
  AND i.approved = 1
  AND i.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND (g.item='board' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = i.imageboard_idimageboard OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(lister_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
LIMIT ? OFFSET ?;

-- name: GetImagePostByIDForLister :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.idimagepost = sqlc.arg(id)
  AND i.approved = 1
  AND i.deleted_at IS NULL
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND (g.item='board' OR g.item IS NULL)
      AND g.action='view'
      AND g.active=1
      AND (g.item_id = i.imageboard_idimageboard OR g.item_id IS NULL)
      AND (g.user_id = sqlc.arg(lister_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
LIMIT 1;

-- name: SystemSetImagePostLastIndex :exec
UPDATE imagepost SET last_index = NOW() WHERE idimagepost = ?;


-- name: GetAllImagePostsForIndex :many
SELECT idimagepost, description FROM imagepost WHERE deleted_at IS NULL;

-- name: GetImagePostInfoByPath :one
SELECT i.idimagepost, i.imageboard_idimageboard, i.users_idusers, i.posted, u.username, b.title
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN imageboard b ON i.imageboard_idimageboard = b.idimageboard
WHERE i.fullimage = ? OR i.thumbnail = ?
LIMIT 1;


-- name: ListImageboardPath :many
WITH RECURSIVE board_path AS (
    SELECT b.idimageboard, b.imageboard_idimageboard AS parent_id, b.title, 0 AS depth
    FROM imageboard b
    WHERE b.idimageboard = sqlc.arg(board_id)
    UNION ALL
    SELECT b2.idimageboard, b2.imageboard_idimageboard, b2.title, p.depth + 1
    FROM imageboard b2
    JOIN board_path p ON p.parent_id = b2.idimageboard
)
SELECT board_path.idimageboard, board_path.title
FROM board_path
ORDER BY board_path.depth DESC;

-- name: AdminListAllImagePosts :many
SELECT i.idimagepost, i.thumbnail, i.fullimage, i.description, i.posted, i.file_size, u.username, b.title
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN imageboard b ON i.imageboard_idimageboard = b.idimageboard
WHERE i.deleted_at IS NULL
ORDER BY i.posted DESC
LIMIT ? OFFSET ?;

-- name: AdminCountAllImagePosts :one
SELECT COUNT(*) FROM imagepost WHERE deleted_at IS NULL;
