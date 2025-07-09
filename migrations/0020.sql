ALTER TABLE uploaded_images
  DROP COLUMN thumbnail,
  ADD COLUMN width INT DEFAULT NULL,
  ADD COLUMN height INT DEFAULT NULL;

UPDATE schema_version SET version = 20 WHERE version = 19;
