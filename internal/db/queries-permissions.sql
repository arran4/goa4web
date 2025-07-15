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
SELECT ur.iduser_roles, ur.users_idusers, r.name AS role
FROM user_roles ur
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
WHERE ur.users_idusers = ? AND r.name = 'administrator';



-- name: GetUsersTopicLevelByUserIdAndThreadId :one
SELECT utl.*
FROM user_topic_permissions utl
WHERE utl.users_idusers = ? AND utl.forumtopic_idforumtopic = ?
;

-- name: DeleteTopicRestrictionsByForumTopicId :exec
DELETE FROM topic_permissions WHERE forumtopic_idforumtopic = ?;

-- name: UpsertForumTopicRestrictions :exec
INSERT INTO topic_permissions (forumtopic_idforumtopic, view_role_id, reply_role_id, newthread_role_id, see_role_id, invite_role_id, read_role_id, mod_role_id, admin_role_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    view_role_id = VALUES(view_role_id),
    reply_role_id = VALUES(reply_role_id),
    newthread_role_id = VALUES(newthread_role_id),
    see_role_id = VALUES(see_role_id),
    invite_role_id = VALUES(invite_role_id),
    read_role_id = VALUES(read_role_id),
    mod_role_id = VALUES(mod_role_id),
    admin_role_id = VALUES(admin_role_id);

-- name: GetForumTopicRestrictionsByForumTopicId :many
SELECT t.idforumtopic, r.*
FROM forumtopic t
LEFT JOIN topic_permissions r ON t.idforumtopic = r.forumtopic_idforumtopic
WHERE idforumtopic = ?;

-- name: GetAllForumTopicRestrictionsWithForumTopicTitle :many
SELECT t.idforumtopic, r.*
FROM forumtopic t
LEFT JOIN topic_permissions r ON t.idforumtopic = r.forumtopic_idforumtopic;


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

-- name: GetPermissionsByUserID :many
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
