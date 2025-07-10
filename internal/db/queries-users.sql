-- name: AllUsers :many
-- This query selects all admin users from the "users" table.
-- Result:
--   idusers (int)
--   username (string)
--   email (string)
SELECT u.idusers, u.username,
       (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email
FROM users u;

-- name: GetUserByUsername :one
SELECT idusers,
       (SELECT email FROM user_emails ue WHERE ue.user_id = users.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email,
       username
FROM users
WHERE username = ?;

-- name: Login :one
SELECT u.idusers,
       (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email,
       p.passwd, p.passwd_algorithm, u.username
FROM users u LEFT JOIN passwords p ON p.users_idusers = u.idusers
WHERE u.username = ?
ORDER BY p.created_at DESC
LIMIT 1;

-- name: GetUserById :one
SELECT u.idusers, ue.email, u.username
FROM users u
LEFT JOIN user_emails ue ON ue.id = (
        SELECT id FROM user_emails ue2
        WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL
        ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE u.idusers = ?;

-- name: UserByUsername :one
SELECT idusers,
       (SELECT email FROM user_emails ue WHERE ue.user_id = users.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email,
       username
FROM users
WHERE username = ?;

-- name: UserByEmail :one
SELECT u.idusers, ue.email, u.username
FROM users u JOIN user_emails ue ON ue.user_id = u.idusers
WHERE ue.email = ?
LIMIT 1;

-- name: InsertUser :execresult
INSERT INTO users (username)
VALUES (?)
;

-- name: ListUsersSubscribedToBlogs :many
SELECT *, (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers ORDER BY ue.id LIMIT 1) AS email
FROM blogs t, users u, preferences p
WHERE t.idblogs=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToNews :many
SELECT idsitenews, forumthread_id, t.language_idlanguage, t.users_idusers,
    news, occurred, u.idusers, u.username, u.deleted_at,
    p.idpreferences, p.language_idlanguage, p.users_idusers, p.emailforumupdates,
    p.page_size, p.auto_subscribe_replies,
    (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers ORDER BY ue.id LIMIT 1) AS email
FROM site_news t, users u, preferences p
WHERE t.idsiteNews=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToLinker :many
SELECT *, (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers ORDER BY ue.id LIMIT 1) AS email
FROM linker t, users u, preferences p
WHERE t.idlinker=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToWriting :many
SELECT *, (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers ORDER BY ue.id LIMIT 1) AS email
FROM writing t, users u, preferences p
WHERE t.idwriting=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=t.users_idusers AND u.idusers!=?
GROUP BY u.idusers;

-- name: ListUsersSubscribedToThread :many
SELECT c.idcomments, c.forumthread_id, c.users_idusers, c.language_idlanguage,
    c.written, c.text, u.idusers,
    (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers ORDER BY ue.id LIMIT 1) AS email,
    u.username,
    p.idpreferences, p.language_idlanguage, p.users_idusers, p.emailforumupdates, p.page_size, p.auto_subscribe_replies
FROM comments c, users u, preferences p
WHERE c.forumthread_id=? AND u.idusers=p.users_idusers AND p.emailforumupdates=1 AND u.idusers=c.users_idusers AND u.idusers!=?
GROUP BY u.idusers;


-- name: ListAdministratorEmails :many
SELECT (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email
FROM users u
JOIN user_roles ur ON ur.users_idusers = u.idusers
JOIN roles r ON ur.role_id = r.id
WHERE ur.section = 'all' and r.name = 'administrator';

-- name: UpdateUserEmail :exec
UPDATE user_emails SET email = ? WHERE user_id = ?;
