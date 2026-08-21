-- name: AdminHardDeleteComment :exec
DELETE FROM comments WHERE idcomments = ?;

-- name: AdminDeleteCommentsByThread :exec
DELETE FROM comments WHERE forumthread_id = ?;

-- name: AdminListBadComments :many
SELECT * FROM comments WHERE text IS NULL OR text = '';
