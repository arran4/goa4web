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
UPDATE notifications
SET read_at = CURRENT_TIMESTAMP
WHERE read_at IS NULL
  AND id IN (
    SELECT ddn.notification_id
    FROM digest_delivery_notifications ddn
    JOIN digest_deliveries dd ON dd.id = ddn.digest_delivery_id
    JOIN pending_emails pe ON pe.digest_delivery_id = dd.id
    WHERE pe.id = sqlc.arg(pending_email_id)
      AND dd.dismiss_notifications = 1
      AND dd.user_id = notifications.users_idusers
  );

-- name: SystemMarkDigestDeliveryAcknowledgedForPendingEmail :exec
UPDATE digest_deliveries
SET acknowledged_at = COALESCE(acknowledged_at, CURRENT_TIMESTAMP)
WHERE digest_deliveries.id = (
  SELECT digest_delivery_id
  FROM pending_emails
  WHERE pending_emails.id = sqlc.arg(pending_email_id)
);
