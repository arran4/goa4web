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

-- name: UpdateNotificationDigestPreferences :exec
UPDATE preferences
SET daily_digest_hour = sqlc.arg(daily_digest_hour),
    daily_digest_mark_read = sqlc.arg(daily_digest_mark_read),
    weekly_digest_day = sqlc.arg(weekly_digest_day),
    weekly_digest_hour = sqlc.arg(weekly_digest_hour),
    monthly_digest_day = sqlc.arg(monthly_digest_day),
    monthly_digest_hour = sqlc.arg(monthly_digest_hour)
WHERE users_idusers = sqlc.arg(lister_id);

-- name: UpdateLastDigestSentAt :exec
UPDATE preferences
SET last_digest_sent_at = sqlc.arg(sent_at)
WHERE users_idusers = sqlc.arg(lister_id);

-- name: UpdateLastWeeklyDigestSentAt :exec
UPDATE preferences
SET last_weekly_digest_sent_at = sqlc.arg(sent_at)
WHERE users_idusers = sqlc.arg(lister_id);

-- name: UpdateLastMonthlyDigestSentAt :exec
UPDATE preferences
SET last_monthly_digest_sent_at = sqlc.arg(sent_at)
WHERE users_idusers = sqlc.arg(lister_id);

-- name: GetUsersForDailyDigest :many
SELECT p.users_idusers, ue.email, p.daily_digest_mark_read
FROM preferences p
JOIN user_emails ue ON ue.id = (
    SELECT id FROM user_emails ue2
    WHERE ue2.user_id = p.users_idusers AND ue2.verified_at IS NOT NULL
    ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE p.daily_digest_hour = sqlc.arg(hour)
  AND (p.last_digest_sent_at IS NULL OR p.last_digest_sent_at < sqlc.arg(cutoff));

-- name: GetPreferenceForLister :one
SELECT idpreferences, language_id, users_idusers, emailforumupdates, page_size, auto_subscribe_replies, timezone, custom_css,
       daily_digest_hour, daily_digest_mark_read, last_digest_sent_at,
       weekly_digest_day, weekly_digest_hour, last_weekly_digest_sent_at,
       monthly_digest_day, monthly_digest_hour, last_monthly_digest_sent_at
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

-- name: GetDigestTimezones :many
SELECT DISTINCT timezone
FROM preferences
WHERE (daily_digest_hour IS NOT NULL OR weekly_digest_day IS NOT NULL OR monthly_digest_day IS NOT NULL)
  AND timezone IS NOT NULL
  AND timezone != '';

-- name: GetUsersForDailyDigestByTimezone :many
SELECT p.users_idusers, ue.email, p.daily_digest_mark_read
FROM preferences p
JOIN user_emails ue ON ue.id = (
    SELECT id FROM user_emails ue2
    WHERE ue2.user_id = p.users_idusers AND ue2.verified_at IS NOT NULL
    ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE p.daily_digest_hour = sqlc.arg(hour)
  AND p.timezone = sqlc.arg(timezone)
  AND (p.last_digest_sent_at IS NULL OR p.last_digest_sent_at < sqlc.arg(cutoff));

-- name: GetUsersForDailyDigestNoTimezone :many
SELECT p.users_idusers, ue.email, p.daily_digest_mark_read
FROM preferences p
JOIN user_emails ue ON ue.id = (
    SELECT id FROM user_emails ue2
    WHERE ue2.user_id = p.users_idusers AND ue2.verified_at IS NOT NULL
    ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE p.daily_digest_hour = sqlc.arg(hour)
  AND (p.timezone IS NULL OR p.timezone = '')
  AND (p.last_digest_sent_at IS NULL OR p.last_digest_sent_at < sqlc.arg(cutoff));

-- name: GetUsersForWeeklyDigestByTimezone :many
SELECT p.users_idusers, ue.email, p.daily_digest_mark_read
FROM preferences p
JOIN user_emails ue ON ue.id = (
    SELECT id FROM user_emails ue2
    WHERE ue2.user_id = p.users_idusers AND ue2.verified_at IS NOT NULL
    ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE p.weekly_digest_day = sqlc.arg(day)
  AND p.weekly_digest_hour = sqlc.arg(hour)
  AND p.timezone = sqlc.arg(timezone)
  AND (p.last_weekly_digest_sent_at IS NULL OR p.last_weekly_digest_sent_at < sqlc.arg(cutoff));

-- name: GetUsersForWeeklyDigestNoTimezone :many
SELECT p.users_idusers, ue.email, p.daily_digest_mark_read
FROM preferences p
JOIN user_emails ue ON ue.id = (
    SELECT id FROM user_emails ue2
    WHERE ue2.user_id = p.users_idusers AND ue2.verified_at IS NOT NULL
    ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE p.weekly_digest_day = sqlc.arg(day)
  AND p.weekly_digest_hour = sqlc.arg(hour)
  AND (p.timezone IS NULL OR p.timezone = '')
  AND (p.last_weekly_digest_sent_at IS NULL OR p.last_weekly_digest_sent_at < sqlc.arg(cutoff));

-- name: GetUsersForMonthlyDigestByTimezone :many
SELECT p.users_idusers, ue.email, p.daily_digest_mark_read
FROM preferences p
JOIN user_emails ue ON ue.id = (
    SELECT id FROM user_emails ue2
    WHERE ue2.user_id = p.users_idusers AND ue2.verified_at IS NOT NULL
    ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE p.monthly_digest_day = sqlc.arg(day)
  AND p.monthly_digest_hour = sqlc.arg(hour)
  AND p.timezone = sqlc.arg(timezone)
  AND (p.last_monthly_digest_sent_at IS NULL OR p.last_monthly_digest_sent_at < sqlc.arg(cutoff));

-- name: GetUsersForMonthlyDigestNoTimezone :many
SELECT p.users_idusers, ue.email, p.daily_digest_mark_read
FROM preferences p
JOIN user_emails ue ON ue.id = (
    SELECT id FROM user_emails ue2
    WHERE ue2.user_id = p.users_idusers AND ue2.verified_at IS NOT NULL
    ORDER BY ue2.notification_priority DESC, ue2.id LIMIT 1
)
WHERE p.monthly_digest_day = sqlc.arg(day)
  AND p.monthly_digest_hour = sqlc.arg(hour)
  AND (p.timezone IS NULL OR p.timezone = '')
  AND (p.last_monthly_digest_sent_at IS NULL OR p.last_monthly_digest_sent_at < sqlc.arg(cutoff));
