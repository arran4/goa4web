-- name: AdminListAllUsers :many
-- Result:
--   idusers (int)
--   username (string)
SELECT DISTINCT u.idusers, u.username
FROM users u
JOIN user_roles ur ON ur.users_idusers = u.idusers
JOIN roles r ON ur.role_id = r.id
WHERE r.is_admin = 1
ORDER BY u.username;

-- name: SystemGetUserByUsername :one
SELECT idusers,
       username,
       public_profile_enabled_at
FROM users
WHERE username = ?;

-- name: SystemGetLogin :one
SELECT u.idusers,
       p.passwd, p.passwd_algorithm, u.username
FROM users u LEFT JOIN passwords p ON p.users_idusers = u.idusers
WHERE u.username = ?
ORDER BY p.created_at DESC
LIMIT 1;

-- name: SystemGetUserByID :one
SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at
FROM users u
LEFT JOIN user_emails ue ON ue.id = (
        SELECT id FROM user_emails ue2
        WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL
        ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE u.idusers = ?;

-- name: SystemGetUserByEmail :one
SELECT u.idusers, ue.email, u.username
FROM users u JOIN user_emails ue ON ue.user_id = u.idusers
WHERE ue.email = ?
LIMIT 1;

-- name: SystemInsertUser :execlastid
INSERT INTO users (username)
VALUES (?);

-- name: SystemListAllUsers :many
SELECT u.idusers, u.username,
       IF(r.id IS NULL, 0, 1) AS admin,
       MIN(s.created_at) AS created_at,
       u.deleted_at
FROM users u
LEFT JOIN user_roles ur ON ur.users_idusers = u.idusers
LEFT JOIN roles r ON ur.role_id = r.id AND r.is_admin = 1
LEFT JOIN sessions s ON s.users_idusers = u.idusers
GROUP BY u.idusers
ORDER BY u.idusers;

-- name: AdminListAdministratorEmails :many
SELECT ue.email
FROM users u
JOIN user_roles ur ON ur.users_idusers = u.idusers
JOIN roles r ON ur.role_id = r.id
JOIN user_emails ue ON ue.user_id = u.idusers AND ue.verified_at IS NOT NULL
WHERE r.is_admin = 1
ORDER BY u.idusers, ue.notification_priority DESC, ue.id;

-- name: AdminUpdateUserEmail :exec
UPDATE user_emails SET email = ? WHERE user_id = ?;

-- name: AdminListPendingUsers :many
SELECT u.idusers, u.username
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

-- name: UpdatePublicProfileEnabledAtForUser :exec
UPDATE users u
SET public_profile_enabled_at = sqlc.arg(enabled_at)
WHERE u.idusers = sqlc.arg(user_id)
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='users'
        AND (g.item='public_profile' OR g.item IS NULL)
        AND g.action='post'
        AND g.active=1
        AND (g.item_id = u.idusers OR g.item_id IS NULL)
        AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (
            SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(user_id)
        ))
  );

-- name: AdminDeleteUserByID :exec
DELETE FROM users WHERE idusers = ?;

-- name: AdminUpdateUsernameByID :exec
UPDATE users SET username = ? WHERE idusers = ?;

-- name: CheckUserHasGrant :one
SELECT EXISTS(
    SELECT 1
    FROM grants g
    WHERE g.user_id = ?
    AND g.section = ?
    AND g.item = ?
    AND g.action = ?
    AND g.active = 1
);

-- name: SystemGetUsersByIDs :many
SELECT u.idusers, ue.email, u.username, u.public_profile_enabled_at
FROM users u
LEFT JOIN user_emails ue ON ue.id = (
        SELECT id FROM user_emails ue2
        WHERE ue2.user_id = u.idusers AND ue2.verified_at IS NOT NULL
        ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE u.idusers IN (sqlc.slice('ids'));
