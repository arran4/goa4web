-- name: InsertPassword :exec
INSERT INTO passwords (users_idusers, passwd, passwd_algorithm)
VALUES (?, ?, ?);

