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

-- name: GetAdministratorPermissionByUserId :one
SELECT *
FROM permissions
WHERE users_idusers = ? AND section = 'all' AND level = 'administrator';

-- name: GetPermissionsByUserIdAndSectionAndSectionAll :one
SELECT *
FROM permissions
WHERE
    users_idusers = ? AND (section = ? OR section = 'all');

-- name: GetPermissionsByUserIdAndSectionBlogs :many
SELECT p.*, u.*
FROM permissions p, users u
WHERE u.idusers = p.users_idusers AND p.section = "blogs"
ORDER BY p.level
;


-- name: GetPermissionsByUserIdAndSectionNews :many
SELECT p.*, u.*
FROM permissions p, users u
WHERE u.idusers = p.users_idusers AND p.section = "news"
ORDER BY p.level
;

-- name: GetPermissionsByUserIdAndSectionWritings :many
SELECT p.*, u.*
FROM permissions p, users u
WHERE u.idusers = p.users_idusers AND (p.section = "writing" OR p.section = "writings")
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
SELECT t.idforumtopic, r.*
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic
WHERE idforumtopic = ?;

-- name: GetAllForumTopicRestrictionsWithForumTopicTitle :many
SELECT t.idforumtopic, r.*
FROM forumtopic t
LEFT JOIN topicrestrictions r ON t.idforumtopic = r.forumtopic_idforumtopic;

-- name: CountPermissionSections :many
SELECT section, COUNT(*) AS SectionCount
FROM permissions
GROUP BY section;

-- name: RenamePermissionSection :exec
UPDATE permissions
SET section = ?
WHERE section = ?;
