-- name: ListRoles :many
SELECT id, name, can_login, is_admin FROM roles ORDER BY id;

-- name: ListRolesWithUsers :many
SELECT r.id, r.name, GROUP_CONCAT(u.username ORDER BY u.username) AS users
FROM roles r
LEFT JOIN user_roles ur ON ur.role_id = r.id
LEFT JOIN users u ON u.idusers = ur.users_idusers
GROUP BY r.id
ORDER BY r.id;
