UPDATE blogs SET forumthread_id = NULL WHERE forumthread_id = 0;

ALTER TABLE blogs
    MODIFY COLUMN forumthread_id int(10) DEFAULT NULL;

-- Update schema version
UPDATE schema_version SET version = 60 WHERE version = 59;
