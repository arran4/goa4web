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
