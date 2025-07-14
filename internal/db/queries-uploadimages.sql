-- name: CreateUploadedImage :execlastid
INSERT INTO uploaded_images (
    users_idusers, path, width, height, file_size, uploaded
) VALUES (?, ?, ?, ?, ?, NOW());

-- name: GetUploadedImage :one
SELECT * FROM uploaded_images WHERE iduploadedimage = ?;

-- name: ListUploadedImagesByUser :many
SELECT iduploadedimage, users_idusers, path, width, height, file_size, uploaded
FROM uploaded_images
WHERE users_idusers = ?
ORDER BY uploaded DESC
LIMIT ? OFFSET ?;
