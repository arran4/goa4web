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
SELECT idpreferences, language_id, users_idusers, emailforumupdates, page_size, auto_subscribe_replies, timezone, custom_css
FROM preferences
WHERE users_idusers = sqlc.arg(lister_id);

-- name: UpdateCustomCssForLister :exec
UPDATE preferences
SET custom_css = sqlc.arg(custom_css)
WHERE users_idusers = sqlc.arg(lister_id);

-- name: InsertPreferenceForLister :exec
INSERT INTO preferences (language_id, users_idusers, page_size, timezone)
VALUES (sqlc.narg(language_id), sqlc.arg(lister_id), sqlc.arg(page_size), sqlc.arg(timezone));

-- name: UpdatePreferenceForLister :exec
UPDATE preferences SET language_id = sqlc.narg(language_id), page_size = sqlc.arg(page_size), timezone = sqlc.arg(timezone) WHERE users_idusers = sqlc.arg(lister_id);

-- name: UpdateTimezoneForLister :exec
UPDATE preferences
SET timezone = sqlc.arg(timezone)
WHERE users_idusers = sqlc.arg(lister_id);
