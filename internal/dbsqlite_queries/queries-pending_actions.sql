-- name: SystemInsertPendingAction :exec
INSERT INTO pending_actions (id, uid, browser_id, action_type, form_data, created_at, expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: SystemGetPendingAction :one
SELECT id, uid, browser_id, action_type, form_data, created_at, expires_at, consumed_at
FROM pending_actions
WHERE id = ? AND expires_at > CURRENT_TIMESTAMP AND consumed_at IS NULL;

-- name: SystemCapturePendingAction :execrows
UPDATE pending_actions
SET form_data = ?, id = ?
WHERE id = ? AND consumed_at IS NULL AND expires_at > CURRENT_TIMESTAMP;

-- name: SystemConsumePendingAction :execrows
UPDATE pending_actions
SET consumed_at = CURRENT_TIMESTAMP
WHERE id = ? AND consumed_at IS NULL AND expires_at > CURRENT_TIMESTAMP;

-- name: SystemDeleteExpiredPendingActions :execrows
DELETE FROM pending_actions
WHERE expires_at <= CURRENT_TIMESTAMP;
