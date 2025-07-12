-- name: CheckRoleGrant :one
SELECT 1
FROM grants g
JOIN roles r ON g.role_id = r.id
WHERE g.section = 'role'
  AND r.name = ?
  AND g.action = ?
  AND g.active = 1
LIMIT 1;
