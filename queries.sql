-- name: renameLanguage :exec
-- This query updates the "nameof" field in the "language" table based on the provided "cid".
-- Parameters:
--   $1 - New name for the language (string)
--   $2 - Language ID to be updated (int)
UPDATE language
SET nameof = $1
WHERE idlanguage = $2;

-- name: deleteLanguage :exec
-- This query deletes a record from the "language" table based on the provided "cid".
-- Parameters:
--   $1 - Language ID to be deleted (int)
DELETE FROM language
WHERE idlanguage = $1;

-- name: countCategories :one
-- This query returns the count of all records in the "language" table.
-- Result:
--   count(*) - The count of rows in the "language" table (int)
SELECT COUNT(*) AS count
FROM language;

-- name: createLanguage :exec
-- This query inserts a new record into the "language" table.
-- Parameters:
--   $1 - Name of the new language (string)
INSERT INTO language (nameof)
VALUES ($1);

-- name: SelectLanguages :many
-- This query selects all languages from the "language" table.
-- Result:
--   idlanguage (int)
--   nameof (string)
SELECT idlanguage, nameof
FROM language;

-- name: adminUserPermissions :many
-- This query selects permissions information for admin users.
-- Result:
--   idpermissions (int)
--   level (int)
--   username (string)
--   email (string)
--   section (string)
SELECT p.idpermissions, p.level, u.username, u.email, p.section
FROM permissions p, users u
WHERE u.idusers = p.users_idusers
ORDER BY p.level;

-- name: userAllow :exec
-- This query inserts a new permission into the "permissions" table.
-- Parameters:
--   $1 - User ID to be associated with the permission (int)
--   $2 - Section for which the permission is granted (string)
--   $3 - Level of the permission (string)
INSERT INTO permissions (users_idusers, section, level)
VALUES ($1, $2, $3);

-- name: userDisallow :exec
-- This query deletes a permission from the "permissions" table based on the provided "permid".
-- Parameters:
--   $1 - Permission ID to be deleted (int)
DELETE FROM permissions
WHERE idpermissions = $1;

-- name: adminUsers :many
-- This query selects all admin users from the "users" table.
-- Result:
--   idusers (int)
--   username (string)
--   email (string)
SELECT u.idusers, u.username, u.email
FROM users u;

-- name: completeWordList :exec
-- This query selects all words from the "searchwordlist" table and prints them.
SELECT word
FROM searchwordlist;

