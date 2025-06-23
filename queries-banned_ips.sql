-- name: InsertBannedIp :exec
INSERT INTO banned_ips (ip_address, reason) VALUES (?, ?);

-- name: DeleteBannedIp :exec
DELETE FROM banned_ips WHERE ip_address = ?;

-- name: GetBannedIpByAddress :one
SELECT * FROM banned_ips WHERE ip_address = ?;

-- name: ListBannedIps :many
SELECT * FROM banned_ips ORDER BY created_at DESC;
