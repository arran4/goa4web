-- name: CreateImageBoard :exec
INSERT INTO imageboard (imageboard_idimageboard, title, description, approval_required) VALUES (?, ?, ?, ?);

-- name: UpdateImageBoard :exec
UPDATE imageboard SET title = ?, description = ?, imageboard_idimageboard = ?, approval_required = ? WHERE idimageboard = ?;

-- name: GetAllBoardsByParentBoardId :many
SELECT *
FROM imageboard
WHERE imageboard_idimageboard = ?;

-- name: GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCount :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.imageboard_idimageboard = ? AND i.approved = 1;

-- name: GetAllImagePostsByIdWithAuthorUsernameAndThreadCommentCount :one
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.idimagepost = ? AND i.approved = 1;

-- name: CreateImagePost :execlastid
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
VALUES (?, ?, ?, ?, ?, NOW(), ?, ?);

-- name: UpdateImagePostByIdForumThreadId :exec
UPDATE imagepost SET forumthread_id = ? WHERE idimagepost = ?;

-- name: GetImagePostsByUserDescending :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.users_idusers = ? AND i.approved = 1
ORDER BY i.posted DESC
LIMIT ? OFFSET ?;

-- name: GetAllImageBoards :many
SELECT b.*
FROM imageboard b
;

-- name: GetImageBoardById :one
SELECT * FROM imageboard WHERE idimageboard = ?;

-- name: DeleteImageBoard :exec
UPDATE imageboard SET deleted_at = NOW() WHERE idimageboard = ?;

-- name: ApproveImagePost :exec
UPDATE imagepost SET approved = 1 WHERE idimagepost = ?;


-- name: GetAllBoardsByParentBoardIdForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT b.*
FROM imageboard b
WHERE b.imageboard_idimageboard = sqlc.arg(parent_id)
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND g.item='board'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = b.idimageboard
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: GetAllImageBoardsForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT b.*
FROM imageboard b
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND g.item='board'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = b.idimageboard
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: GetImagePostsByUserDescendingForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.users_idusers = sqlc.arg(user_id)
  AND i.approved = 1
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND g.item='board'
      AND g.action='see'
      AND g.active=1
      AND g.item_id = i.imageboard_idimageboard
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY i.posted DESC
LIMIT ? OFFSET ?;

-- name: GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUser :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.imageboard_idimageboard = sqlc.arg(board_id)
  AND i.approved = 1
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND g.item='board'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = i.imageboard_idimageboard
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  );

-- name: GetAllImagePostsByIdWithAuthorUsernameAndThreadCommentCountForUser :one
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_id = th.idforumthread
WHERE i.idimagepost = sqlc.arg(id)
  AND i.approved = 1
  AND EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='imagebbs'
      AND g.item='board'
      AND g.action='view'
      AND g.active=1
      AND g.item_id = i.imageboard_idimageboard
      AND (g.user_id = sqlc.arg(viewer_user_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
LIMIT 1;

-- name: SetImagePostLastIndex :exec
UPDATE imagepost SET last_index = NOW() WHERE idimagepost = ?;


-- name: GetAllImagePostsForIndex :many
SELECT idimagepost, description FROM imagepost WHERE deleted_at IS NULL;

