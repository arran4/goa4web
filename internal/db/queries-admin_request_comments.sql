-- name: InsertAdminRequestComment :exec
INSERT INTO admin_request_comments (request_id, comment)
VALUES (?, ?);

-- name: ListAdminRequestComments :many
SELECT id, request_id, comment, created_at
FROM admin_request_comments
WHERE request_id = ?
ORDER BY id DESC;
