-- name: UpdateEmailForumUpdatesForLister :exec
UPDATE preferences
SET emailforumupdates = sqlc.arg(email_forum_updates)
WHERE users_idusers = sqlc.arg(lister_id);

-- name: InsertEmailPreferenceForLister :exec
INSERT INTO preferences (emailforumupdates, auto_subscribe_replies, users_idusers)
VALUES (sqlc.arg(email_forum_updates), sqlc.arg(auto_subscribe_replies), sqlc.arg(lister_id));

-- name: UpdateAutoSubscribeRepliesForLister :exec
UPDATE preferences
SET auto_subscribe_replies = sqlc.arg(auto_subscribe_replies)
WHERE users_idusers = sqlc.arg(lister_id);

-- name: GetPreferenceForLister :one
SELECT idpreferences, language_idlanguage, users_idusers, emailforumupdates, page_size, auto_subscribe_replies
FROM preferences
WHERE users_idusers = sqlc.arg(lister_id);

-- name: InsertPreferenceForLister :exec
INSERT INTO preferences (language_idlanguage, users_idusers, page_size)
VALUES (sqlc.arg(language_id), sqlc.arg(lister_id), sqlc.arg(page_size));

-- name: UpdatePreferenceForLister :exec
UPDATE preferences SET language_idlanguage = sqlc.arg(language_id), page_size = sqlc.arg(page_size) WHERE users_idusers = sqlc.arg(lister_id);
