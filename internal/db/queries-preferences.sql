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
