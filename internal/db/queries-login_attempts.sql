-- name: InsertLoginAttempt :exec
INSERT INTO login_attempts (username, ip_address)
VALUES (?, ?);

-- name: ListLoginAttempts :many
SELECT *
FROM login_attempts
ORDER BY id DESC;


-- name: CountRecentLoginAttempts :one
SELECT COUNT(*) FROM login_attempts
WHERE (username = ? OR ip_address = ?) AND created_at > ?;
