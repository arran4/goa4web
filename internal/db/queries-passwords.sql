-- name: InsertPassword :exec
INSERT INTO passwords (users_idusers, passwd, passwd_algorithm)
VALUES (?, ?, ?);

-- name: GetLatestPasswordByUserID :one
SELECT id, users_idusers, passwd, passwd_algorithm, created_at
FROM passwords
WHERE users_idusers = ?
ORDER BY created_at DESC
LIMIT 1;
