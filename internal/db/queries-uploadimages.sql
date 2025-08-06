-- name: CreateUploadedImageForUploader :execlastid
INSERT INTO uploaded_images (
    users_idusers, path, width, height, file_size, uploaded
)
SELECT sqlc.arg(uploader_id), sqlc.arg(path), sqlc.arg(width), sqlc.arg(height), sqlc.arg(file_size), NOW()
WHERE EXISTS (
    SELECT 1 FROM grants g
    WHERE g.section='images'
      AND g.item='upload'
      AND g.action='post'
      AND g.active=1
      AND (g.user_id = sqlc.arg(grantee_id) OR g.user_id IS NULL)
      AND (g.role_id IS NULL OR g.role_id IN (
          SELECT ur.role_id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(uploader_id)
      ))
);

-- name: ListUploadedImagesByUserForLister :many
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = sqlc.arg(lister_id)
)
SELECT ui.iduploadedimage, ui.users_idusers, ui.path, ui.width, ui.height, ui.file_size, ui.uploaded
FROM uploaded_images ui
WHERE ui.users_idusers = sqlc.arg(user_id)
  AND EXISTS (
      SELECT 1 FROM grants g
      WHERE g.section='images'
        AND g.item='upload'
        AND g.action='see'
        AND g.active=1
        AND (g.user_id = sqlc.arg(lister_match_id) OR g.user_id IS NULL)
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
