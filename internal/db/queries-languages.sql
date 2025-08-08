-- AdminRenameLanguage updates the language name.
-- Parameters:
--   ? - New name for the language (string)
--   ? - Language ID to be updated (int)
-- name: AdminRenameLanguage :exec
UPDATE language
SET nameof = ?
WHERE idlanguage = ?;

-- AdminDeleteLanguage removes a language entry.
-- Parameters:
--   ? - Language ID to be deleted (int)
-- name: AdminDeleteLanguage :exec
DELETE FROM language
WHERE idlanguage = ?;

-- AdminCreateLanguage adds a new language.
-- Parameters:
--   ? - Name of the new language (string)
-- name: AdminCreateLanguage :exec
INSERT INTO language (nameof)
VALUES (?);

-- AdminInsertLanguage adds a new language returning a result.
-- name: AdminInsertLanguage :execresult
INSERT INTO language (nameof)
VALUES (?);

-- SystemListLanguages lists all languages.
-- name: SystemListLanguages :many
SELECT *
FROM language;

-- SystemGetLanguageIDByName resolves a language ID by name.
-- name: SystemGetLanguageIDByName :one
SELECT idlanguage FROM language WHERE nameof = ?;


-- SystemCountLanguages counts all languages.
-- name: SystemCountLanguages :one
SELECT COUNT(*) FROM language;

-- AdminLanguageUsageCounts returns counts of content referencing a language.
-- name: AdminLanguageUsageCounts :one
SELECT
    (SELECT COUNT(*) FROM comments WHERE comments.language_idlanguage = sqlc.narg(lang_id)) AS comments,
    (SELECT COUNT(*) FROM writing WHERE writing.language_idlanguage = sqlc.narg(lang_id)) AS writings,
    (SELECT COUNT(*) FROM blogs WHERE blogs.language_idlanguage = sqlc.narg(lang_id)) AS blogs,
    (SELECT COUNT(*) FROM site_news WHERE site_news.language_idlanguage = sqlc.narg(lang_id)) AS news,
    (SELECT COUNT(*) FROM linker WHERE linker.language_idlanguage = sqlc.narg(lang_id)) AS links;
