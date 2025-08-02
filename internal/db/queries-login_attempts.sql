-- name: SystemInsertLoginAttempt :exec
INSERT INTO login_attempts (username, ip_address)
VALUES (?, ?);

-- name: AdminListLoginAttempts :many
SELECT id, username, ip_address, created_at
FROM login_attempts
ORDER BY id DESC;


-- name: SystemCountRecentLoginAttempts :one
SELECT COUNT(*) FROM login_attempts
WHERE (username = ? OR ip_address = ?) AND created_at > ?;
