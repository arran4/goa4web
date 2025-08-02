-- name: AdminInsertRequestQueue :execresult
INSERT INTO admin_request_queue (users_idusers, change_table, change_field, change_row_id, change_value, contact_options)
VALUES (?, ?, ?, ?, ?, ?);

-- name: AdminListPendingRequests :many
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE status = 'pending'
ORDER BY id;

-- name: AdminListArchivedRequests :many
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE status <> 'pending'
ORDER BY id DESC;

-- name: AdminGetRequestByID :one
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE id = ?;

-- name: AdminUpdateRequestStatus :exec
UPDATE admin_request_queue SET status = ?, acted_at = NOW() WHERE id = ?;
