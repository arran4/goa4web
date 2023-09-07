-- name: CreateImageBoard :exec
INSERT INTO imageboard (imageboard_idimageboard, title, description) VALUES (?, ?, ?);

-- name: UpdateImageBoard :exec
UPDATE imageboard SET title = ?, description = ?, imageboard_idimageboard = ? WHERE idimageboard = ?;

-- name: GetAllBoardsByParentBoardId :many
SELECT *
FROM imageboard
WHERE imageboard_idimageboard = ?;

-- name: GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCount :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.imageboard_idimageboard = ?;

-- name: GetAllImagePostsByIdWithAuthorUsernameAndThreadCommentCount :one
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.idimagepost = ?;

-- name: CreateImagePost :exec
INSERT INTO imagepost (imageboard_idimageboard, thumbnail, fullimage, users_idusers, description, posted)
VALUES (?, ?, ?, ?, ?, NOW());

-- name: UpdateImagePostByIdForumThreadId :exec
UPDATE imagepost SET forumthread_idforumthread = ? WHERE idimagepost = ?;

-- name: GetAllImageBoards :many
SELECT b.*
FROM imageboard b
;

