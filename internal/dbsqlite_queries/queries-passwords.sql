-- name: InsertPassword :exec
INSERT INTO passwords (users_idusers, passwd, passwd_algorithm)
VALUES (?, ?, ?);

-- name: GetPendingPassword :one
SELECT * FROM pending_passwords WHERE user_id = ?;

-- name: GetPendingPasswordByCode :one
SELECT * FROM pending_passwords WHERE verification_code = ?;

-- name: UpdateUserPassword :exec
INSERT INTO passwords (users_idusers, passwd, passwd_algorithm) VALUES (?, ?, ?) ON CONFLICT(users_idusers) DO UPDATE SET passwd = excluded.passwd, passwd_algorithm = excluded.passwd_algorithm;

-- name: DeletePendingPassword :exec
DELETE FROM pending_passwords WHERE user_id = ?;
