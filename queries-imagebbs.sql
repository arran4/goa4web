-- name: ImageboardRSS :exec
SELECT title, description FROM imageboard WHERE idimageboard = ?;

-- name: MakeImageBoard :exec
INSERT INTO imageboard (imageboard_idimageboard, title, description) VALUES (?, ?, ?);

-- name: ChangeImageBoard :exec
UPDATE imageboard SET title = ?, description = ?, imageboard_idimageboard = ? WHERE idimageboard = ?;

-- name: PrintSubBoards :many
SELECT idimageboard, title, description FROM imageboard WHERE imageboard_idimageboard = ?;

-- name: PrintImagePosts :many
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.imageboard_idimageboard = ?;

-- name: PrintImagePost :one
SELECT i.*, u.username, th.comments
FROM imagepost i
LEFT JOIN users u ON i.users_idusers = u.idusers
LEFT JOIN forumthread th ON i.forumthread_idforumthread = th.idforumthread
WHERE i.idimagepost = ?;

-- name: AddImage :exec
INSERT INTO imagepost (imageboard_idimageboard, thumbnail, fullimage, users_idusers, description, posted)
VALUES (?, ?, ?, ?, ?, NOW());

-- name: AssignImagePostThisThreadId :exec
UPDATE imagepost SET forumthread_idforumthread = ? WHERE idimagepost = ?;

-- name: ShowAllBoards :many
SELECT b.idimageboard, b.title, b.description, b.imageboard_idimageboard, pb.title
FROM imageboard b
LEFT JOIN imageboard pb ON b.imageboard_idimageboard = pb.idimageboard OR b.imageboard_idimageboard = 0
GROUP BY b.idimageboard;

