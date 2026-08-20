-- +goose Up
ALTER TABLE blogs RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE comments RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE imagepost RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE linker RENAME COLUMN linkerCategory_idlinkerCategory TO linker_category_id;
ALTER TABLE linker RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE linker_queue RENAME COLUMN linkerCategory_idlinkerCategory TO linker_category_id;
ALTER TABLE site_news RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE writing RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE deactivated_comments RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE deactivated_writings RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE deactivated_blogs RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE deactivated_imageposts RENAME COLUMN forumthread_idforumthread TO forumthread_id;
ALTER TABLE deactivated_linker RENAME COLUMN linkerCategory_idlinkerCategory TO linker_category_id;
ALTER TABLE deactivated_linker RENAME COLUMN forumthread_idforumthread TO forumthread_id;
UPDATE schema_version SET version = 25;