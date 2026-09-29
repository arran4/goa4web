-- name: InsertPendingAction :exec
INSERT INTO pending_actions (id, uid, browser_id, action_type, form_data, created_at, expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetPendingAction :one
SELECT id, uid, browser_id, action_type, form_data, created_at, expires_at, consumed_at
FROM pending_actions
WHERE id = ? AND expires_at > CURRENT_TIMESTAMP AND consumed_at IS NULL;

-- name: UpdatePendingActionData :execrows
UPDATE pending_actions
SET form_data = ?, id = ?
WHERE id = ? AND consumed_at IS NULL AND expires_at > CURRENT_TIMESTAMP;

-- name: ConsumePendingAction :execrows
UPDATE pending_actions
SET consumed_at = CURRENT_TIMESTAMP
WHERE id = ? AND consumed_at IS NULL AND expires_at > CURRENT_TIMESTAMP;

-- name: DeletePendingActionsForUser :exec
DELETE FROM pending_actions
WHERE uid = ?;
