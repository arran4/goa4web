ALTER TABLE linker
    CHANGE COLUMN idlinker id INT NOT NULL AUTO_INCREMENT,
    CHANGE COLUMN language_idlanguage language_id INT(10) DEFAULT NULL,
    CHANGE COLUMN users_idusers author_id INT(10) NOT NULL DEFAULT 0,
    CHANGE COLUMN linker_category_id category_id INT(10) DEFAULT NULL,
    CHANGE COLUMN forumthread_id thread_id INT(10) NOT NULL DEFAULT 0;

ALTER TABLE linker_category
    CHANGE COLUMN idlinkerCategory id INT(10) NOT NULL AUTO_INCREMENT;

ALTER TABLE linker_queue
    CHANGE COLUMN idlinkerQueue id INT(10) NOT NULL AUTO_INCREMENT,
    CHANGE COLUMN language_idlanguage language_id INT(10) DEFAULT NULL,
    CHANGE COLUMN users_idusers submitter_id INT(10) NOT NULL DEFAULT 0,
    CHANGE COLUMN linker_category_id category_id INT(10) DEFAULT NULL;

-- Update schema version
UPDATE schema_version SET version = 66 WHERE version = 65;
