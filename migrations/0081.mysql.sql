ALTER TABLE external_links ADD COLUMN card_duration TINYTEXT;
ALTER TABLE external_links ADD COLUMN card_upload_date TINYTEXT;
ALTER TABLE external_links ADD COLUMN card_author TINYTEXT;

UPDATE uploaded_images SET path = TRIM(LEADING '/' FROM path);
UPDATE uploaded_images SET path = TRIM(LEADING 'uploads' FROM path);
UPDATE uploaded_images SET path = TRIM(LEADING '/' FROM path);

INSERT INTO schema_version (version) VALUES (81) ON DUPLICATE KEY UPDATE version = VALUES(version);
