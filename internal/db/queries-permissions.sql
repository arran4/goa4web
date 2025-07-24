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
--   email (string)
SELECT ur.iduser_roles, ur.users_idusers, r.name AS role,
       u.username,
       (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers ORDER BY ue.id LIMIT 1) AS email
FROM user_roles ur
JOIN users u ON u.idusers = ur.users_idusers
JOIN roles r ON ur.role_id = r.id
;

-- name: CreateUserRole :exec
-- This query inserts a new permission into the "permissions" table.
-- Parameters:
--   ? - User ID to be associated with the permission (int)
--   ? - Role of the permission (string)
INSERT INTO user_roles (users_idusers, role_id)
SELECT ?, r.id FROM roles r WHERE r.name = ?;

-- name: DeleteUserRole :exec
-- This query deletes a permission from the "permissions" table based on the provided "permid".
-- Parameters:
--   ? - Permission ID to be deleted (int)
DELETE FROM user_roles
WHERE iduser_roles = ?;

-- name: GetAdministratorUserRole :one
SELECT ur.*
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ? AND r.is_admin = 1;





-- name: CheckRoleGrant :one
SELECT 1
FROM grants g
JOIN roles r ON g.role_id = r.id
WHERE g.section = 'role'
  AND r.name = ?
  AND g.action = ?
  AND g.active = 1
LIMIT 1;

-- name: ListEffectiveRoleIDsByUserID :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = ?
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT DISTINCT id FROM role_ids;

-- name: CheckGrant :one
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT 1 FROM grants g
WHERE g.section = sqlc.arg(section)
  AND (g.item = sqlc.arg(item) OR g.item IS NULL)
  AND g.action = sqlc.arg(action)
  AND g.active = 1
  AND (g.item_id = sqlc.arg(item_id) OR g.item_id IS NULL)
  AND (g.user_id = sqlc.arg(user_id) OR g.user_id IS NULL)
  AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
LIMIT 1;

-- name: CreateGrant :execlastid
INSERT INTO grants (
    created_at, user_id, role_id, section, item, rule_type, item_id, item_rule, action, extra, active
) VALUES (NOW(), ?, ?, ?, ?, ?, ?, ?, ?, ?, 1);

-- name: DeleteGrant :exec
DELETE FROM grants WHERE id = ?;

-- name: ListGrants :many
SELECT * FROM grants ORDER BY id;

-- name: UserHasRole :one
SELECT 1
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ? AND r.name = ?
LIMIT 1;

-- name: UserHasLoginRole :one
SELECT 1
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ? AND r.can_login = 1
LIMIT 1;

-- name: GetPermissionsByUserID :many
-- Lists the role names granted to a user.
SELECT ur.iduser_roles, ur.users_idusers, r.name
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ?;

-- name: GetPermissionsWithUsers :many
SELECT ur.iduser_roles, ur.users_idusers, r.name, u.username,
       (SELECT email FROM user_emails ue WHERE ue.user_id = u.idusers ORDER BY ue.id LIMIT 1) AS email
FROM user_roles ur
JOIN users u ON u.idusers = ur.users_idusers
JOIN roles r ON ur.role_id = r.id
WHERE (sqlc.arg(username) = '' OR u.username = sqlc.arg(username));

-- name: UpdatePermission :exec
UPDATE user_roles SET role_id = (SELECT id FROM roles WHERE name = ?) WHERE iduser_roles = ?;
-- name: ListUsersWithRoles :many
SELECT u.idusers, u.username, GROUP_CONCAT(r.name ORDER BY r.name) AS roles
FROM users u
LEFT JOIN user_roles ur ON u.idusers = ur.users_idusers
LEFT JOIN roles r ON r.id = ur.role_id
GROUP BY u.idusers
ORDER BY u.idusers;
