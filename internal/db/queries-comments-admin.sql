-- name: AdminHardDeleteComment :exec
DELETE comments, comments_search
FROM comments
LEFT JOIN comments_search ON comments_search.comment_id = comments.idcomments
WHERE comments.idcomments = ?;

-- name: AdminDeleteCommentsByThread :exec
DELETE comments, comments_search
FROM comments
LEFT JOIN comments_search ON comments_search.comment_id = comments.idcomments
WHERE comments.forumthread_id = ?;

-- name: AdminListBadComments :many
SELECT * FROM comments WHERE text IS NULL OR text = '';
