-- admin task
-- name: AdminRenameLanguage :exec
-- This query updates the "nameof" field in the "language" table based on the provided "cid".
-- Parameters:
--   ? - New name for the language (string)
--   ? - Language ID to be updated (int)
UPDATE language
SET nameof = ?
WHERE idlanguage = ?;

-- admin task
-- name: AdminDeleteLanguage :exec
-- This query deletes a record from the "language" table based on the provided "cid".
-- Parameters:
--   ? - Language ID to be deleted (int)
DELETE FROM language
WHERE idlanguage = ?;


-- admin task
-- name: AdminInsertLanguage :execresult
INSERT INTO language (nameof)
VALUES (?);

-- user listing
-- name: ListLanguagesForUser :many
SELECT idlanguage, nameof
FROM language
WHERE NOT EXISTS (
    SELECT 1 FROM user_language ul WHERE ul.users_idusers = sqlc.arg(viewer_id)
) OR EXISTS (
    SELECT 1 FROM user_language ul
    WHERE ul.users_idusers = sqlc.arg(viewer_id)
      AND ul.language_idlanguage = idlanguage
)
ORDER BY nameof;

-- admin task
-- name: AdminListLanguages :many
SELECT idlanguage, nameof FROM language
ORDER BY nameof;

-- name: GetLanguageIDByName :one
SELECT idlanguage FROM language WHERE nameof = ?;


-- name: CountLanguages :one
SELECT COUNT(*) FROM language;
