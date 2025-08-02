-- name: AdminInsertBannedIp :exec
INSERT INTO banned_ips (ip_net, reason, expires_at)
VALUES (?, ?, ?);

-- name: AdminUpdateBannedIp :exec
UPDATE banned_ips SET reason = ?, expires_at = ? WHERE id = ?;

-- name: AdminCancelBannedIp :exec
UPDATE banned_ips SET canceled_at = CURRENT_TIMESTAMP WHERE ip_net = ? AND canceled_at IS NULL;


-- name: ListActiveBans :many
SELECT * FROM banned_ips WHERE canceled_at IS NULL AND (expires_at IS NULL OR expires_at > NOW());

-- name: ListBannedIps :many
SELECT * FROM banned_ips ORDER BY created_at DESC;
