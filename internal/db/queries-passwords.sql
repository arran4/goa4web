-- name: InsertPassword :exec
INSERT INTO passwords (users_idusers, passwd, passwd_algorithm)
VALUES (?, ?, ?);

-- name: GetPendingPassword :one
SELECT * FROM pending_passwords WHERE user_id = ?;

-- name: UpdateUserPassword :exec
INSERT INTO passwords (users_idusers, passwd, passwd_algorithm) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE passwd = VALUES(passwd), passwd_algorithm = VALUES(passwd_algorithm);

-- name: DeletePendingPassword :exec
DELETE FROM pending_passwords WHERE user_id = ?;
