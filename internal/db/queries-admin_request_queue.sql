-- name: AdminInsertRequestQueue :execresult
INSERT INTO admin_request_queue (users_idusers, change_table, change_field, change_row_id, change_value, contact_options)
VALUES (?, ?, ?, ?, ?, ?);

-- name: AdminListRequestQueue :many
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: AdminListRequestQueueByStatus :many
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE status = ?
ORDER BY created_at ASC;

-- name: AdminGetRequest :one
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE id = ?;

-- name: AdminGetRequestByID :one
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE id = ?;

-- name: AdminUpdateRequestStatus :exec
UPDATE admin_request_queue
SET status = ?, acted_at = NOW()
WHERE id = ?;

-- name: AdminListRequestsByUserID :many
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE users_idusers = ?
ORDER BY created_at DESC;

-- name: AdminUpdateRequestStatusByTableAndRow :exec
UPDATE admin_request_queue
SET status = ?, acted_at = NOW()
WHERE change_table = ? AND change_row_id = ? AND status = 'pending';

-- name: AdminListPendingRequests :many
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: AdminListArchivedRequests :many
SELECT id, users_idusers, change_table, change_field, change_row_id, change_value, contact_options, status, created_at, acted_at
FROM admin_request_queue
WHERE status != 'pending'
ORDER BY acted_at DESC;
