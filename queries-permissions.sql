-- name: GetUserPermissions :one
-- This query selects permissions information for admin users.
-- Result:
--   idpermissions (int)
--   level (int)
--   username (string)
--   email (string)
--   section (string)
SELECT p.*
FROM permissions p
WHERE p.users_idusers = ?
;

-- name: GetUsersPermissions :many
-- This query selects permissions information for admin users.
-- Result:
--   idpermissions (int)
--   level (int)
--   username (string)
--   email (string)
--   section (string)
SELECT p.*
FROM permissions p
;

-- name: PermissionUserAllow :exec
-- This query inserts a new permission into the "permissions" table.
-- Parameters:
--   ? - User ID to be associated with the permission (int)
--   ? - Section for which the permission is granted (string)
--   ? - Level of the permission (string)
INSERT INTO permissions (users_idusers, section, level)
VALUES (?, ?, ?);

-- name: PermissionUserDisallow :exec
-- This query deletes a permission from the "permissions" table based on the provided "permid".
-- Parameters:
--   ? - Permission ID to be deleted (int)
DELETE FROM permissions
WHERE idpermissions = ?;

-- name: GetPermissionsByUserIdAndSectionAndSectionAll :one
SELECT level FROM permissions WHERE users_idusers = ? AND (section = ? OR section = 'all');

-- name: GetPermissionsByUserIdAndSectionBlogs :many
SELECT p.idpermissions, p.level, u.username, u.email, p.section
FROM permissions p, users u
WHERE u.idusers = p.users_idusers AND p.section = "blogs"
ORDER BY p.level
;

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
SELECT idforumtopic, r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, t.title, r.forumtopic_idforumtopic, r.modlevel, r.adminlevel
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
WHERE idforumtopic = ?;

-- name: GetAllForumTopicRestrictionsWithForumTopicTitle :many
SELECT idforumtopic, r.viewlevel, r.replylevel, r.newthreadlevel, r.seelevel, r.invitelevel, r.readlevel, t.title, r.forumtopic_idforumtopic, r.modlevel, r.adminlevel
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic;

