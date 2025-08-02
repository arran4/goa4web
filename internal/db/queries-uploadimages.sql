-- name: CreateUploadedImage :execlastid
INSERT INTO uploaded_images (
    users_idusers, path, width, height, file_size, uploaded
) VALUES (?, ?, ?, ?, ?, NOW());

-- name: ListUploadedImagesByUserForViewer :many
WITH RECURSIVE role_ids(id) AS (
    SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(viewer_id)
    UNION
    SELECT r2.id
    FROM role_ids ri
    JOIN grants g ON g.role_id = ri.id AND g.section = 'role' AND g.active = 1
    JOIN roles r2 ON r2.name = g.action
)
SELECT ui.iduploadedimage, ui.users_idusers, ui.path, ui.width, ui.height, ui.file_size, ui.uploaded
FROM uploaded_images ui
WHERE ui.users_idusers = sqlc.arg(user_id)
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='images'
        AND (g.item='upload' OR g.item IS NULL)
        AND g.action='see'
        AND g.active=1
        AND (g.user_id = sqlc.arg(viewer_match_id) OR g.user_id IS NULL)
        AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
  )
ORDER BY uploaded DESC
LIMIT ? OFFSET ?;

-- name: AdminListUploadedImages :many
-- Admin
SELECT iduploadedimage, users_idusers, path, width, height, file_size, uploaded
FROM uploaded_images
ORDER BY uploaded DESC
LIMIT ? OFFSET ?;
