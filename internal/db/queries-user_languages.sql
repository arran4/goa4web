-- name: GetUserLanguages :many
SELECT iduser_language, users_idusers, language_idlanguage
FROM user_language
WHERE users_idusers = ?;

-- name: DeleteUserLanguagesByUser :exec
DELETE FROM user_language WHERE users_idusers = ?;

-- name: InsertUserLang :exec
INSERT INTO user_language (users_idusers, language_idlanguage)
VALUES (?, ?);
