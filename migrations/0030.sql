ALTER TABLE blogs
    MODIFY COLUMN forumthread_id int(10) DEFAULT NULL;

-- Record upgrade to schema version 30
UPDATE schema_version SET version = 30 WHERE version = 29;
