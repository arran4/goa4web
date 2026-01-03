-- name: AdminListRoles :many
-- admin task
SELECT id, name, can_login, is_admin, private_labels, public_profile_allowed_at FROM roles ORDER BY id;

-- name: AdminListRolesWithUsers :many
-- admin task
SELECT r.id, r.name, GROUP_CONCAT(u.username ORDER BY u.username) AS users
FROM roles r
LEFT JOIN user_roles ur ON ur.role_id = r.id
LEFT JOIN users u ON u.idusers = ur.users_idusers
GROUP BY r.id
ORDER BY r.id;

-- name: AdminUpdateRolePublicProfileAllowed :exec
-- admin task
UPDATE roles SET public_profile_allowed_at = ? WHERE id = ?;

-- name: AdminGetRoleByID :one
-- admin task
SELECT id, name, can_login, is_admin, private_labels, public_profile_allowed_at FROM roles WHERE id = ?;

-- name: AdminListUsersByRoleID :many
-- admin task
SELECT u.idusers, u.username
FROM users u
JOIN user_roles ur ON ur.users_idusers = u.idusers
WHERE ur.role_id = ?
ORDER BY u.username;

-- name: AdminListGrantsByRoleID :many
-- admin task
SELECT * FROM grants WHERE role_id = ? ORDER BY id;

-- name: AdminUpdateRole :exec
-- admin task
UPDATE roles SET name = ?, can_login = ?, is_admin = ?, private_labels = ?, public_profile_allowed_at = ? WHERE id = ?;

-- name: GetRoleByName :one
SELECT id, name, can_login, is_admin, private_labels, public_profile_allowed_at FROM roles WHERE name = ?;

-- name: GetGrantsByRoleID :many
SELECT * FROM grants WHERE role_id = ?;

-- name: CreateGrant :exec
INSERT INTO grants (role_id, section, item, rule_type, item_id, item_rule, action, extra, active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteGrantsByRoleID :exec
DELETE FROM grants WHERE role_id = ?;

-- name: DeleteGrantByProperties :exec
DELETE FROM grants
WHERE role_id = ?
  AND section = ?
  AND item = ?
  AND action = ?;
