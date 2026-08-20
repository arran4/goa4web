-- name: InsertAuditLog :exec
INSERT INTO audit_log (users_idusers, action, path, details, data) VALUES (?, ?, ?, ?, ?);

-- name: AdminListAuditLogs :many
SELECT a.id, a.users_idusers, a.action, a.path, a.details, a.data, a.created_at, u.username
FROM audit_log a
LEFT JOIN users u ON a.users_idusers = idusers
WHERE u.username LIKE sqlc.arg(username)
  AND a.action LIKE sqlc.arg(action)
  AND a.path LIKE sqlc.arg(section)
  AND (sqlc.arg(start_time) IS NULL OR a.created_at >= sqlc.arg(start_time))
  AND (sqlc.arg(end_time) IS NULL OR a.created_at <= sqlc.arg(end_time))
ORDER BY a.id DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: AdminAuditLogActionSummary :many
SELECT a.action, COUNT(*) AS total
FROM audit_log a
LEFT JOIN users u ON a.users_idusers = idusers
WHERE u.username LIKE sqlc.arg(username)
  AND a.action LIKE sqlc.arg(action)
  AND a.path LIKE sqlc.arg(section)
  AND (sqlc.arg(start_time) IS NULL OR a.created_at >= sqlc.arg(start_time))
  AND (sqlc.arg(end_time) IS NULL OR a.created_at <= sqlc.arg(end_time))
GROUP BY a.action
ORDER BY total DESC, a.action ASC;
