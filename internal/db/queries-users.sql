-- name: AllUsers :many
-- This query selects all admin users from the "users" table.
-- Result:
--   idusers (int)
--   username (string)
--   email (string)
SELECT u.*
FROM users u;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = ?;

-- name: Login :one
SELECT *
FROM users
WHERE username = ? AND passwd = md5(?);

-- name: GetUserById :one
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

-- name: ListUsersSubscribedToBlogs :many
SELECT *
FROM blogs t, users u, preferences p
WHERE t.idblogs=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToLinker :many
SELECT *
FROM linker t, users u, preferences p
WHERE t.idlinker=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToWriting :many
SELECT *
FROM writing t, users u, preferences p
WHERE t.idwriting=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToThread :many
SELECT *
FROM comments c, users u, preferences p
WHERE c.forumthread_idforumthread=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=c.users_idusers AND u.idusers!=?
GROUP BY u.idusers;


-- name: ListAdministratorEmails :many
SELECT u.email
FROM users u
JOIN permissions p ON p.users_idusers = u.idusers
WHERE p.section = 'administrator';

-- name: UpdateUserEmail :exec
UPDATE users
SET email = ?
WHERE idusers = ?;
