-- name: InsertBannedIpForAdmin :exec
INSERT INTO banned_ips (ip_net, reason, expires_at)
VALUES (?, ?, ?);

-- name: UpdateBannedIpForAdmin :exec
UPDATE banned_ips SET reason = ?, expires_at = ? WHERE id = ?;

-- name: CancelBannedIpForAdmin :exec
UPDATE banned_ips SET canceled_at = CURRENT_TIMESTAMP WHERE ip_net = ? AND canceled_at IS NULL;

-- name: GetBannedIpByAddressForAdmin :one
SELECT * FROM banned_ips WHERE ip_net = ?;

-- name: ListActiveBansSystem :many
SELECT * FROM banned_ips WHERE canceled_at IS NULL AND (expires_at IS NULL OR expires_at > NOW());

-- name: ListBannedIpsForAdmin :many
SELECT * FROM banned_ips ORDER BY created_at DESC;
