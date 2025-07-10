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
-- Result:
--   idpermissions (int)
--   role (string)
--   username (string)
--   email (string)
SELECT ur.idpermissions, ur.users_idusers, r.name AS role
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
WHERE idpermissions = ?;

-- name: GetAdministratorUserRole :one
SELECT ur.*
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.users_idusers = ? AND r.name = 'administrator';



-- name: GetUsersTopicLevelByUserIdAndThreadId :one
SELECT utl.*
FROM userstopiclevel utl
WHERE utl.users_idusers = ? AND utl.forumtopic_idforumtopic = ?
;

-- name: DeleteTopicRestrictionsByForumTopicId :exec
DELETE FROM topicrestrictions WHERE forumtopic_idforumtopic = ?;

-- name: UpsertForumTopicRestrictions :exec
INSERT INTO topicrestrictions (forumtopic_idforumtopic, viewlevel, replylevel, newthreadlevel, seelevel, invitelevel, readlevel, modlevel, adminlevel)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    viewlevel = VALUES(viewlevel),
    replylevel = VALUES(replylevel),
    newthreadlevel = VALUES(newthreadlevel),
    seelevel = VALUES(seelevel),
    invitelevel = VALUES(invitelevel),
    readlevel = VALUES(readlevel),
    modlevel = VALUES(modlevel),
    adminlevel = VALUES(adminlevel);

-- name: GetForumTopicRestrictionsByForumTopicId :many
SELECT t.idforumtopic, r.*
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
WHERE idforumtopic = ?;

-- name: GetAllForumTopicRestrictionsWithForumTopicTitle :many
SELECT t.idforumtopic, r.*
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic;

