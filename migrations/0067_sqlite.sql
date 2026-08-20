-- +goose Up
ALTER TABLE linker RENAME COLUMN idlinker TO id;
ALTER TABLE linker RENAME COLUMN users_idusers TO author_id;
ALTER TABLE linker RENAME COLUMN linker_category_id TO category_id;
ALTER TABLE linker RENAME COLUMN forumthread_id TO thread_id;

ALTER TABLE linker_category RENAME COLUMN idlinkerCategory TO id;

ALTER TABLE linker_queue RENAME COLUMN idlinkerQueue TO id;
ALTER TABLE linker_queue RENAME COLUMN users_idusers TO submitter_id;
ALTER TABLE linker_queue RENAME COLUMN linker_category_id TO category_id;

ALTER TABLE deactivated_linker RENAME COLUMN idlinker TO id;
ALTER TABLE deactivated_linker RENAME COLUMN users_idusers TO author_id;
ALTER TABLE deactivated_linker RENAME COLUMN linker_category_id TO category_id;
ALTER TABLE deactivated_linker RENAME COLUMN forumthread_id TO thread_id;

UPDATE schema_version SET version = 67;