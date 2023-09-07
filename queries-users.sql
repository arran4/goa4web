-- name: AllUsers :many
-- This query selects all admin users from the "users" table.
-- Result:
--   idusers (int)
--   username (string)
--   email (string)
SELECT u.*
FROM users u;

-- name: Usernametouid :one
SELECT idusers FROM users WHERE username = ?;

-- name: Login :one
SELECT *
FROM users
WHERE username = ? AND passwd = md5(?);

-- name: UserByUid :one
SELECT *
FROM users
WHERE idusers = ?;

-- name: UserByUsername :one
SELECT *
FROM users
WHERE username = ?;

-- name: UserByEmail :one
SELECT *
FROM users
WHERE email = ?;

-- name: InsertUser :execresult
INSERT INTO users (username, passwd, email)
VALUES (?, MD5(?), ?)
;

