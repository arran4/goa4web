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
