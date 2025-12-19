-- name: AdminHardDeleteComment :exec
DELETE FROM comments WHERE idcomments = ?;

-- name: AdminListBadComments :many
SELECT * FROM comments WHERE text IS NULL OR text = '';
