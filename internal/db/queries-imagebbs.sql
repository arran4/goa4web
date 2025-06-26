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
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.imageboard_idimageboard = ? AND i.approved = 1;

-- name: GetAllImagePostsByIdWithAuthorUsernameAndThreadCommentCount :one
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
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
UPDATE imagepost SET forumthread_idforumthread = ? WHERE idimagepost = ?;

-- name: GetImagePostsByUserDescending :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
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

