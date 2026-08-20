-- name: InsertSubscription :exec
INSERT INTO subscriptions (users_idusers, pattern, method)
VALUES (?, ?, ?);

-- name: DeleteSubscriptionForSubscriber :exec
DELETE FROM subscriptions
WHERE users_idusers = sqlc.arg(subscriber_id) AND pattern = sqlc.arg(pattern) AND method = sqlc.arg(method);

-- name: ListSubscribersForPattern :many
SELECT users_idusers FROM subscriptions
WHERE pattern = ? AND method = ?;

-- name: ListSubscribersForPatterns :many
SELECT DISTINCT users_idusers FROM subscriptions
WHERE pattern IN (sqlc.slice(patterns)) AND method = ?;

-- name: ListSubscriptionsByUser :many
SELECT id, pattern, method FROM subscriptions
WHERE users_idusers = ?
ORDER BY id;

-- name: ListThreadSubscriptionsByUser :many
SELECT id, pattern, method FROM subscriptions
WHERE users_idusers = ?
  AND pattern LIKE 'reply:/forum/topic/%/thread/%'
  AND pattern NOT LIKE '%/topic/*/%'
  AND pattern NOT LIKE '%/thread/*%'
ORDER BY id;

-- name: DeleteSubscriptionByIDForSubscriber :exec
DELETE FROM subscriptions WHERE users_idusers = sqlc.arg(subscriber_id) AND id = sqlc.arg(id);

-- name: UpdateSubscriptionByIDForSubscriber :exec
UPDATE subscriptions SET pattern = sqlc.arg(pattern), method = sqlc.arg(method)
WHERE users_idusers = sqlc.arg(subscriber_id) AND id = sqlc.arg(id);
