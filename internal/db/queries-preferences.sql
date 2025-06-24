-- name: UpdateEmailForumUpdatesByUserID :exec
UPDATE preferences
SET emailforumupdates = ?
WHERE users_idusers = ?;

-- name: InsertEmailPreference :exec
INSERT INTO preferences (emailforumupdates, users_idusers)
VALUES (?, ?);
