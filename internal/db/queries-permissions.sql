-- name: AdminCreateGrant :execlastid
INSERT INTO grants (
    created_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active
) VALUES (NOW(), ?, ?, ?, ?, ?, ?, ?, ?, ?, 1);

-- name: AdminDeleteGrant :exec
DELETE FROM grants WHERE id = ?;

-- name: AdminUpdateGrantActive :exec
UPDATE grants SET active = ? WHERE id = ?;

-- name: AdminGetRoleByNameForUser :one
SELECT 1
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ? AND r.name = ?
LIMIT 1;

-- name: GetAdministratorUserRole :one
SELECT ur.*
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ? AND r.is_admin = 1;

-- name: GetUserRole :one
-- This query returns the role for a user.
-- Result:
--   role (string)
SELECT r.name as role
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ?
LIMIT 1;

-- name: GetUserRoles :many
-- This query selects permissions information for admin users.
--   iduser_roles (int)
--   role (string)
--   username (string)
SELECT ur.iduser_roles, ur.users_idusers, r.name AS role, r.is_admin,
       u.username
FROM user_roles ur
JOIN users u ON u.idusers = ur.users_idusers
JOIN roles r ON ur.role_id = r.id
ORDER BY ur.users_idusers;

-- name: ListUsersWithRoles :many
SELECT u.idusers, u.username, GROUP_CONCAT(r.name ORDER BY r.name) AS roles
FROM users u
LEFT JOIN user_roles ur ON u.idusers = ur.users_idusers
LEFT JOIN roles r ON r.id = ur.role_id
GROUP BY u.idusers
ORDER BY u.idusers;

-- name: GetPublicProfileRoleForUser :one
SELECT 1
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ? AND r.public_profile_allowed_at IS NOT NULL
LIMIT 1;

-- name: GetLoginRoleForUser :one
SELECT 1
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ? AND r.can_login = 1
LIMIT 1;

-- name: SystemCheckGrant :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT id FROM roles WHERE name = 'anyone'
)
SELECT 1 FROM grants g
WHERE g.section = ?
  AND (g.item = ? OR g.item IS NULL)
  AND g.action = ?
  AND g.active = 1
  AND (g.item_id = ? OR g.item_id IS NULL)
  AND (g.user_id = ? OR g.user_id IS NULL)
  AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
LIMIT 1;

-- name: SystemCheckRoleGrant :one
SELECT 1
FROM grants g
JOIN roles r ON g.role_id = r.id
WHERE g.section = 'role'
  AND r.name = ?
  AND g.action = ?
  AND g.active = 1
LIMIT 1;

-- name: ListGrants :many
SELECT * FROM grants ORDER BY id;

-- name: ListGrantsExtended :many
SELECT g.*, u.username, r.name as role_name
FROM grants g
LEFT JOIN users u ON g.user_id = u.idusers
LEFT JOIN roles r ON g.role_id = r.id
WHERE
    (sqlc.narg('user_id') IS NULL OR g.user_id = sqlc.narg('user_id'))
AND (sqlc.narg('username') IS NULL OR u.username = sqlc.narg('username'))
AND (sqlc.narg('role_id') IS NULL OR g.role_id = sqlc.narg('role_id'))
AND (sqlc.narg('role_name') IS NULL OR r.name = sqlc.narg('role_name'))
AND (sqlc.arg('only_roles') = false OR g.role_id IS NOT NULL)
AND (sqlc.arg('only_users') = false OR g.user_id IS NOT NULL)
ORDER BY g.id;

-- name: ListGrantsByUserID :many
SELECT * FROM grants WHERE user_id = ? ORDER BY id;

-- name: GetPermissionsByUserID :many
-- Lists the role names granted to a user.
SELECT ur.iduser_roles, ur.users_idusers, ur.role_id, r.name, r.is_admin
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ?;

-- name: GetPermissionsWithUsers :many
SELECT ur.iduser_roles, ur.users_idusers, r.name, u.username
FROM user_roles ur
JOIN users u ON u.idusers = ur.users_idusers
JOIN roles r ON ur.role_id = r.id
WHERE (sqlc.arg(username) = '' OR u.username = sqlc.arg(username));

-- name: SystemCreateGrant :execlastid
INSERT INTO grants (
    created_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active
) VALUES (NOW(), ?, ?, ?, ?, ?, ?, ?, ?, ?, 1);

-- name: SystemCreateUserRole :exec
-- This query inserts a new permission into the "permissions" table.
-- Parameters:
--   ? - User ID to be associated with the permission (int)
--   ? - Role of the permission (string)
INSERT INTO user_roles (users_idusers, role_id)
SELECT ?, r.id FROM roles r WHERE r.name = ?;

-- name: SystemCreateUserRoleByID :exec
-- This query inserts a new permission into the "permissions" table by role ID.
-- Parameters:
--   ? - User ID to be associated with the permission (int)
--   ? - Role ID (int)
INSERT INTO user_roles (users_idusers, role_id)
VALUES (?, ?);

-- name: AdminUpdateUserRole :exec
UPDATE user_roles SET role_id = (SELECT id FROM roles WHERE name = ?) WHERE iduser_roles = ?;

-- name: AdminDeleteUserRole :exec
-- This query deletes a permission from the "permissions" table based on the provided "permid".
-- Parameters:
--   ? - Permission ID to be deleted (int)
DELETE FROM user_roles
WHERE iduser_roles = ?;

-- name: ListEffectiveRoleIDsByUserID :many
SELECT DISTINCT ur.role_id AS id
FROM user_roles ur
WHERE ur.users_idusers = ?;

-- name: GetVerifiedUserEmails :many
-- Fetch verified (active) email addresses ordered by notification priority.
SELECT ue.user_id, ue.email
FROM user_emails ue
WHERE ue.verified_at IS NOT NULL
ORDER BY ue.user_id, ue.notification_priority DESC, ue.id;

-- name: SearchGrants :many
SELECT g.*, u.username, r.name as role_name
FROM grants g
LEFT JOIN users u ON g.user_id = u.idusers
LEFT JOIN roles r ON g.role_id = r.id
WHERE
    (sqlc.narg('section') IS NULL OR g.section = sqlc.narg('section'))
AND (sqlc.narg('item') IS NULL OR g.item = sqlc.narg('item'))
AND (sqlc.narg('item_id') IS NULL OR g.item_id = sqlc.narg('item_id'))
AND (sqlc.narg('active') IS NULL OR g.active = sqlc.narg('active'))
AND (sqlc.narg('username') IS NULL OR u.username LIKE sqlc.narg('username'))
AND (sqlc.narg('role_name') IS NULL OR r.name LIKE sqlc.narg('role_name'))
ORDER BY g.id;

-- name: GetUserRoleByID :one
-- Gets a specific user role assignment by its primary key (iduser_roles).
SELECT ur.iduser_roles, ur.users_idusers, r.name, u.username
FROM user_roles ur
JOIN users u ON u.idusers = ur.users_idusers
JOIN roles r ON ur.role_id = r.id
WHERE ur.iduser_roles = sqlc.arg('id');
