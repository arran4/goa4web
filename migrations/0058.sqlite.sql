ALTER TABLE writing_category
    MODIFY COLUMN writing_category_id INT DEFAULT NULL;
UPDATE writing_category SET writing_category_id = NULL WHERE writing_category_id = 0;

-- Record upgrade to schema version 58
UPDATE schema_version SET version = 58 WHERE version = 57;
