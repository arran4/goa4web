-- name: AdminInsertUserComment :exec
INSERT INTO admin_user_comments (users_idusers, comment)
VALUES (?, ?);

-- name: AdminListUserComments :many
SELECT id, users_idusers, comment, created_at
FROM admin_user_comments
WHERE users_idusers = ?
ORDER BY id DESC;

-- name: AdminGetLatestUserComment :one
SELECT id, users_idusers, comment, created_at
FROM admin_user_comments
WHERE users_idusers = ?
ORDER BY id DESC
LIMIT 1;
