UPDATE uploaded_images SET path = TRIM(LEADING '/' FROM path);
UPDATE uploaded_images SET path = TRIM(LEADING 'uploads' FROM path);
UPDATE uploaded_images SET path = TRIM(LEADING '/' FROM path);
