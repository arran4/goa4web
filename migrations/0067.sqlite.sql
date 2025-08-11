ALTER TABLE linker
    CHANGE COLUMN idlinker id INT NOT NULL AUTO_INCREMENT,
    CHANGE COLUMN users_idusers author_id INT NOT NULL DEFAULT 0,
    CHANGE COLUMN linker_category_id category_id INT NULL DEFAULT NULL,
    CHANGE COLUMN forumthread_id thread_id INT NOT NULL DEFAULT 0;

ALTER TABLE linker_category
    CHANGE COLUMN idlinkerCategory id INT NOT NULL AUTO_INCREMENT;

ALTER TABLE linker_queue
    CHANGE COLUMN idlinkerQueue id INT NOT NULL AUTO_INCREMENT,
    CHANGE COLUMN users_idusers submitter_id INT NOT NULL DEFAULT 0,
    CHANGE COLUMN linker_category_id category_id INT NULL DEFAULT NULL;

ALTER TABLE linker RENAME COLUMN idlinker TO id;
ALTER TABLE linker RENAME COLUMN users_idusers TO author_id;
ALTER TABLE linker RENAME COLUMN linker_category_id TO category_id;
ALTER TABLE linker RENAME COLUMN forumthread_id TO thread_id;

ALTER TABLE linker_category RENAME COLUMN idlinkerCategory TO id;

ALTER TABLE linker_queue RENAME COLUMN idlinkerQueue TO id;
ALTER TABLE linker_queue RENAME COLUMN users_idusers TO submitter_id;
ALTER TABLE linker_queue RENAME COLUMN linker_category_id TO category_id;

-- Update schema version
UPDATE schema_version SET version = 67 WHERE version = 66;
