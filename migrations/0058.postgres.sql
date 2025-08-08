ALTER TABLE writing_category
    ALTER COLUMN writing_category_id DROP NOT NULL;
UPDATE writing_category SET writing_category_id = NULL WHERE writing_category_id = 0;

-- Record upgrade to schema version 58
UPDATE schema_version SET version = 58 WHERE version = 57;
