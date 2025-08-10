ALTER TABLE writing_category ALTER COLUMN writing_category_id DROP DEFAULT;
ALTER TABLE writing_category ALTER COLUMN writing_category_id DROP NOT NULL;

UPDATE writing_category SET writing_category_id = NULL WHERE writing_category_id = 0;

-- Update schema version
UPDATE schema_version SET version = 59 WHERE version = 58;
