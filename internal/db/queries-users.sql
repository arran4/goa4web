-- name: AdminListAllUsers :many
-- Result:
--   idusers (int)
--   username (string)
--   email (string)
SELECT u.idusers, u.username,
       (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email
FROM users u
JOIN user_roles ur ON ur.users_idusers = u.idusers
JOIN roles r ON ur.role_id = r.id
WHERE r.is_admin = 1;

-- name: GetUserByUsername :one
SELECT idusers,
       (SELECT email FROM user_emails ue WHERE ue.user_id = users.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email,
       username,
       public_profile_enabled_at
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
SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at
FROM users u
LEFT JOIN user_emails ue ON ue.id = (
        SELECT id FROM user_emails ue2
        WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL
        ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE u.idusers = ?;

-- name: UserByEmail :one
SELECT u.idusers, ue.email, u.username
FROM users u JOIN user_emails ue ON ue.user_id = u.idusers
WHERE ue.email = ?
LIMIT 1;

-- name: InsertUser :execresult
INSERT INTO users (username)
VALUES (?)
;

-- name: AdminListAdministratorEmails :many
SELECT (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email
FROM users u
JOIN user_roles ur ON ur.users_idusers = u.idusers
JOIN roles r ON ur.role_id = r.id
WHERE r.is_admin = 1;

-- name: UpdateUserEmail :exec
UPDATE user_emails SET email = ? WHERE user_id = ?;

-- name: AdminListPendingUsers :many
SELECT u.idusers, u.username,
       (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers AND ue.verified_at IS NOT NULL ORDER BY ue.notification_priority DESC, ue.id LIMIT 1) AS email
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.users_idusers = u.idusers AND (r.can_login = 1 OR r.name = 'rejected')
)
ORDER BY u.idusers;


-- name: AdminListUserIDsByRole :many
SELECT u.idusers
FROM users u
JOIN user_roles ur ON ur.users_idusers = u.idusers
JOIN roles r ON ur.role_id = r.id
WHERE r.name = ?
ORDER BY u.idusers;

-- name: AdminListAllUserIDs :many
SELECT idusers FROM users ORDER BY idusers;

-- name: AdminListUsersByID :many
SELECT idusers, username
FROM users
WHERE idusers IN (sqlc.slice('ids'));

-- name: UpdatePublicProfileEnabledAtByUserID :exec
UPDATE users SET public_profile_enabled_at = ? WHERE idusers = ?;
