-- name: InsertAuditLog :exec
INSERT INTO audit_log (users_idusers, action) VALUES (?, ?);

-- name: ListAuditLogs :many
SELECT a.id, a.users_idusers, a.action, a.created_at, u.username
FROM audit_log a
LEFT JOIN users u ON a.users_idusers = u.idusers
WHERE u.username LIKE sqlc.arg(username) AND a.action LIKE sqlc.arg(action)
ORDER BY a.id DESC
LIMIT ? OFFSET ?;
