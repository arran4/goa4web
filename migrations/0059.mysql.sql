ALTER TABLE writing_category
    MODIFY COLUMN writing_category_id int(10) DEFAULT NULL;

UPDATE writing_category SET writing_category_id = NULL WHERE writing_category_id = 0;

-- Update schema version
UPDATE schema_version SET version = 59 WHERE version = 58;
