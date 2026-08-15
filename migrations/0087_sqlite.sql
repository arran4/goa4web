-- +goose Up
UPDATE uploaded_images
SET path = CASE
    WHEN path LIKE '/uploads/%' THEN SUBSTRING(path, 9)
    WHEN path LIKE 'uploads/%' THEN CONCAT('/', SUBSTRING(path, 9))
    ELSE path
END
WHERE path LIKE '/uploads/%' OR path LIKE 'uploads/%';