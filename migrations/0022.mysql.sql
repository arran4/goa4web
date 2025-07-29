RENAME TABLE writingCategory TO writing_category;

-- Record upgrade to schema version 22
UPDATE schema_version SET version = 22 WHERE version = 21;
