-- name: InsertAdminUserComment :exec
INSERT INTO admin_user_comments (users_idusers, comment)
VALUES (?, ?);

-- name: ListAdminUserComments :many
SELECT id, users_idusers, comment, created_at
FROM admin_user_comments
WHERE users_idusers = ?
ORDER BY id DESC;

-- name: LatestAdminUserComment :one
SELECT id, users_idusers, comment, created_at
FROM admin_user_comments
WHERE users_idusers = ?
ORDER BY id DESC
LIMIT 1;
