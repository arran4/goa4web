ALTER TABLE blogs
    CHANGE COLUMN forumthread_idforumthread forumthread_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE comments
    CHANGE COLUMN forumthread_idforumthread forumthread_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE imagepost
    CHANGE COLUMN forumthread_idforumthread forumthread_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE linker
    CHANGE COLUMN linkerCategory_idlinkerCategory linker_category_id int(10) NOT NULL DEFAULT 0,
    CHANGE COLUMN forumthread_idforumthread forumthread_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE linker_queue
    CHANGE COLUMN linkerCategory_idlinkerCategory linker_category_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE site_news
    CHANGE COLUMN forumthread_idforumthread forumthread_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE writing
    CHANGE COLUMN forumthread_idforumthread forumthread_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE deactivated_comments
    CHANGE COLUMN forumthread_idforumthread forumthread_id int NOT NULL;
ALTER TABLE deactivated_writings
    CHANGE COLUMN forumthread_idforumthread forumthread_id int NOT NULL;
ALTER TABLE deactivated_blogs
    CHANGE COLUMN forumthread_idforumthread forumthread_id int NOT NULL;
ALTER TABLE deactivated_imageposts
    CHANGE COLUMN forumthread_idforumthread forumthread_id int NOT NULL;
ALTER TABLE deactivated_linker
    CHANGE COLUMN linkerCategory_idlinkerCategory linker_category_id int NOT NULL,
    CHANGE COLUMN forumthread_idforumthread forumthread_id int NOT NULL;

-- Record upgrade to schema version 25
UPDATE schema_version SET version = 25 WHERE version = 24;
