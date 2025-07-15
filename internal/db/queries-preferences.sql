-- name: UpdateEmailForumUpdatesByUserID :exec
UPDATE preferences
SET emailforumupdates = ?
WHERE users_idusers = ?;

-- name: InsertEmailPreference :exec
INSERT INTO preferences (emailforumupdates, auto_subscribe_replies, users_idusers)
VALUES (?, ?, ?);

-- name: UpdateAutoSubscribeRepliesByUserID :exec
UPDATE preferences
SET auto_subscribe_replies = ?
WHERE users_idusers = ?;

-- name: GetPreferenceByUserID :one
SELECT idpreferences, language_idlanguage, users_idusers, emailforumupdates, page_size, auto_subscribe_replies
FROM preferences
WHERE users_idusers = ?;

-- name: InsertPreference :exec
INSERT INTO preferences (language_idlanguage, users_idusers, page_size)
VALUES (?, ?, ?);

-- name: UpdatePreference :exec
UPDATE preferences SET language_idlanguage = ?, page_size = ? WHERE users_idusers = ?;
