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

