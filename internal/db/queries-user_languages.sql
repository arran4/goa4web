-- name: GetUserLanguages :many
SELECT iduserlang, users_idusers, language_idlanguage
FROM user_language
WHERE users_idusers = ?;

-- name: DeleteUserLanguagesForUser :exec
DELETE FROM user_language WHERE users_idusers = sqlc.arg(user_id);

-- name: InsertUserLang :exec
INSERT INTO user_language (users_idusers, language_idlanguage)
VALUES (?, ?);
