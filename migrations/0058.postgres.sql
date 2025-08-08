ALTER TABLE linker
    MODIFY COLUMN linker_category_id int(10) DEFAULT NULL;
ALTER TABLE linker_queue
    MODIFY COLUMN linker_category_id int(10) DEFAULT NULL;
ALTER TABLE deactivated_linker
    MODIFY COLUMN linker_category_id int DEFAULT NULL;

UPDATE linker SET linker_category_id = NULL WHERE linker_category_id = 0;
UPDATE linker_queue SET linker_category_id = NULL WHERE linker_category_id = 0;
UPDATE deactivated_linker SET linker_category_id = NULL WHERE linker_category_id = 0;

UPDATE schema_version SET version = 58 WHERE version = 57;
