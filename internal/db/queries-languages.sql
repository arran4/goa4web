-- name: RenameLanguage :exec
-- This query updates the "nameof" field in the "language" table based on the provided "cid".
-- Parameters:
--   ? - New name for the language (string)
--   ? - Language ID to be updated (int)
UPDATE language
SET nameof = ?
WHERE idlanguage = ?;

-- name: DeleteLanguage :exec
-- This query deletes a record from the "language" table based on the provided "cid".
-- Parameters:
--   ? - Language ID to be deleted (int)
DELETE FROM language
WHERE idlanguage = ?;

-- name: CreateLanguage :exec
-- This query inserts a new record into the "language" table.
-- Parameters:
--   ? - Name of the new language (string)
INSERT INTO language (nameof)
VALUES (?);

-- name: InsertLanguage :execresult
INSERT INTO language (nameof)
VALUES (?);

-- name: FetchLanguages :many
SELECT *
FROM language;

-- name: AllLanguages :many
SELECT * FROM language;

-- name: GetLanguageIDByName :one
SELECT idlanguage FROM language WHERE nameof = ?;

