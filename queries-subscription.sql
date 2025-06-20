-- name: CreateSubscription :exec
INSERT INTO subscriptions (users_idusers, forumthread_idforumthread)
VALUES (?, ?) ON DUPLICATE KEY UPDATE users_idusers=users_idusers;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions WHERE users_idusers = ? AND forumthread_idforumthread = ?;

-- name: ListSubscribersForThread :many
SELECT u.username
FROM subscriptions s
JOIN users u ON s.users_idusers = u.idusers
JOIN preferences p ON u.idusers = p.users_idusers
WHERE s.forumthread_idforumthread = ? AND p.emailforumupdates = 1 AND u.idusers != ?;
