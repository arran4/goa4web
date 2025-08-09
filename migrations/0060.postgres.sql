UPDATE blogs SET forumthread_id = NULL WHERE forumthread_id = 0;

ALTER TABLE blogs
    ALTER COLUMN forumthread_id DROP DEFAULT,
    ALTER COLUMN forumthread_id DROP NOT NULL;

-- Update schema version
UPDATE schema_version SET version = 60 WHERE version = 59;
