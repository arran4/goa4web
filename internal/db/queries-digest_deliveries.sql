-- name: SystemCreateDigestDelivery :execlastid
INSERT INTO digest_deliveries (user_id, dismiss_notifications)
VALUES (?, ?);

-- name: SystemCreateDigestDeliveryNotification :exec
INSERT INTO digest_delivery_notifications (digest_delivery_id, notification_id)
VALUES (?, ?);

-- name: SystemInsertDigestPendingEmail :exec
INSERT INTO pending_emails (to_user_id, body, direct_email, digest_delivery_id)
VALUES (?, ?, ?, ?);

-- name: SystemAcknowledgeDigestNotificationsForPendingEmail :exec
UPDATE notifications n
JOIN digest_delivery_notifications ddn ON ddn.notification_id = n.id
JOIN digest_deliveries dd ON dd.id = ddn.digest_delivery_id
JOIN pending_emails pe ON pe.digest_delivery_id = dd.id
SET n.read_at = NOW()
WHERE pe.id = ?
  AND dd.dismiss_notifications = 1
  AND n.users_idusers = dd.user_id
  AND n.read_at IS NULL;

-- name: SystemMarkDigestDeliveryAcknowledgedForPendingEmail :exec
UPDATE digest_deliveries dd
JOIN pending_emails pe ON pe.digest_delivery_id = dd.id
SET dd.acknowledged_at = COALESCE(dd.acknowledged_at, NOW())
WHERE pe.id = ?;
