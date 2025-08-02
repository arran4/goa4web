-- name: AdminInsertBannedIp :exec
-- admin task
INSERT INTO banned_ips (ip_net, reason, expires_at)
VALUES (?, ?, ?);

-- name: AdminUpdateBannedIp :exec
-- admin task
UPDATE banned_ips SET reason = ?, expires_at = ? WHERE id = ?;

-- name: AdminCancelBannedIp :exec
-- admin task
UPDATE banned_ips SET canceled_at = CURRENT_TIMESTAMP WHERE ip_net = ? AND canceled_at IS NULL;


-- name: SystemListActiveBans :many
SELECT ip_net
FROM banned_ips
WHERE canceled_at IS NULL AND (expires_at IS NULL OR expires_at > NOW());

-- name: AdminListBannedIps :many
-- admin task
SELECT id, ip_net, reason, created_at, expires_at, canceled_at
FROM banned_ips
ORDER BY created_at DESC;
