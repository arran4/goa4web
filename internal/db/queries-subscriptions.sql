-- name: InsertSubscription :exec
INSERT INTO subscriptions (users_idusers, pattern, method)
VALUES (?, ?, ?);

-- name: DeleteSubscription :exec
DELETE FROM subscriptions
WHERE users_idusers = ? AND pattern = ? AND method = ?;

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

-- name: DeleteSubscriptionByID :exec
DELETE FROM subscriptions WHERE users_idusers = ? AND id = ?;
