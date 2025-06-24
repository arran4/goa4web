-- name: InsertSession :exec
INSERT INTO sessions (session_id, users_idusers)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE users_idusers = VALUES(users_idusers);

-- name: DeleteSessionByID :exec
DELETE FROM sessions WHERE session_id = ?;

-- name: ListSessions :many
SELECT s.session_id, s.users_idusers, u.username
FROM sessions s
LEFT JOIN users u ON u.idusers = s.users_idusers
ORDER BY s.session_id;
