-- Administrative creation of a new image board.
-- Admin commands must not require a user ID.
-- name: AdminCreateImageBoard :exec
INSERT INTO imageboard (imageboard_idimageboard, title, description, approval_required)
VALUES (
    sqlc.arg(parent_id),
    sqlc.arg(title),
    sqlc.arg(description),
    sqlc.arg(approval_required)
);

-- Administrative update of board properties.
-- name: AdminUpdateImageBoard :exec
UPDATE imageboard
SET
    title = sqlc.arg(title),
    description = sqlc.arg(description),
    imageboard_idimageboard = sqlc.arg(parent_id),
    approval_required = sqlc.arg(approval_required)
WHERE idimageboard = sqlc.arg(board_id);

-- name: SystemListBoardsByParentID :many
SELECT b.*
FROM imageboard b
WHERE b.imageboard_idimageboard = sqlc.arg(parent_id)
LIMIT ? OFFSET ?;


-- Posts an image on behalf of a poster.
-- name: CreateImagePostForPoster :execlastid
INSERT INTO imagepost (
    imageboard_idimageboard,
    thumbnail,
    fullimage,
    users_idusers,
    description,
    posted,
    approved,
    file_size
)
VALUES (
    sqlc.arg(board_id),
    sqlc.arg(thumbnail),
    sqlc.arg(fullimage),
    sqlc.arg(poster_id),
    sqlc.arg(description),
    NOW(),
    sqlc.arg(approved),
    sqlc.arg(file_size)
);

-- Administrative association of an image post to a forum thread.
-- name: AdminSetImagePostForumThreadID :exec
UPDATE imagepost
SET forumthread_id = sqlc.arg(forum_thread_id)
WHERE idimagepost = sqlc.arg(image_post_id);

-- Administrative listing of approved image posts by poster.
-- name: AdminListImagePostsByPoster :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.users_idusers = sqlc.arg(poster_id) AND i.approved = 1 AND i.deleted_at IS NULL
ORDER BY i.posted DESC
LIMIT ? OFFSET ?;

-- Administrative listing of all image posts by poster regardless of approval.
-- name: AdminListAllImagePostsByPoster :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.users_idusers = sqlc.arg(poster_id) AND i.deleted_at IS NULL
ORDER BY i.posted DESC
LIMIT ? OFFSET ?;

-- name: AdminListBoards :many
SELECT b.*
FROM imageboard b
LIMIT ? OFFSET ?;

-- Administrative fetch of a board by ID.
-- name: AdminGetImageBoardByID :one
SELECT * FROM imageboard WHERE idimageboard = sqlc.arg(board_id);

-- Administrative soft delete of a board.
-- name: AdminDeleteImageBoard :exec
UPDATE imageboard SET deleted_at = NOW() WHERE idimageboard = sqlc.arg(board_id);

-- Administrative approval of an image post.
-- name: AdminApproveImagePost :exec
UPDATE imagepost SET approved = 1 WHERE idimagepost = sqlc.arg(image_post_id);

-- name: GetImageBoardByIDForLister :one
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT b.*
FROM imageboard b
WHERE b.idimageboard = sqlc.arg(board_id)
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
  );


-- name: ListBoardsByParentIDForLister :many
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT b.*
FROM imageboard b
WHERE b.imageboard_idimageboard = sqlc.arg(parent_id)
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
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT b.*
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
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
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
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
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
      AND g.item='board'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = i.imageboard_idimageboard
      AND (g.user_id = sqlc.arg(lister_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
LIMIT ? OFFSET ?;

-- name: GetImagePostByIDForLister :one
WITH RECURSIVE role_ids(id) AS (
    SELECT DISTINCT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
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
      AND g.item='board'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = i.imageboard_idimageboard
      AND (g.user_id = sqlc.arg(lister_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
LIMIT 1;

-- name: SetImagePostLastIndex :exec
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


