-- name: CreateUploadedImage :execlastid
INSERT INTO uploaded_images (
    users_idusers, path, width, height, file_size, uploaded
) VALUES (?, ?, ?, ?, ?, NOW());

-- name: GetUploadedImage :one
SELECT * FROM uploaded_images WHERE iduploadedimage = ?;
