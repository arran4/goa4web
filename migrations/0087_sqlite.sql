-- +goose Up
UPDATE uploaded_images
SET path = CASE
    WHEN path LIKE '/uploads/%' THEN substr(path, 9)
    WHEN path LIKE 'uploads/%' THEN '/' || substr(path, 9)
    ELSE path
END
WHERE path LIKE '/uploads/%' OR path LIKE 'uploads/%';
UPDATE schema_version SET version = 87;