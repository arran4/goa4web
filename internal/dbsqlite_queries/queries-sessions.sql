-- name: SystemInsertSession :exec
INSERT INTO sessions (session_id, users_idusers)
VALUES (?, ?)
ON CONFLICT(session_id) DO UPDATE SET users_idusers = excluded.users_idusers;

-- name: SystemDeleteSessionByID :exec
DELETE FROM sessions WHERE session_id = ?;

-- name: AdminListSessions :many
SELECT session_id, s.users_idusers, u.username
FROM sessions s
LEFT JOIN users u ON idusers = s.users_idusers
ORDER BY session_id;
