-- name: InsertLoginAttempt :exec
INSERT INTO login_attempts (username, ip_address)
VALUES (?, ?);

-- name: ListLoginAttempts :many
SELECT *
FROM login_attempts
ORDER BY id DESC;

