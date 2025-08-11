ALTER TABLE linker RENAME COLUMN idlinker TO id;
ALTER TABLE linker RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE linker RENAME COLUMN users_idusers TO author_id;
ALTER TABLE linker RENAME COLUMN linker_category_id TO category_id;
ALTER TABLE linker RENAME COLUMN forumthread_id TO thread_id;

ALTER TABLE linker_category RENAME COLUMN idlinkerCategory TO id;

ALTER TABLE linker_queue RENAME COLUMN idlinkerQueue TO id;
ALTER TABLE linker_queue RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE linker_queue RENAME COLUMN users_idusers TO submitter_id;
ALTER TABLE linker_queue RENAME COLUMN linker_category_id TO category_id;

-- Update schema version
UPDATE schema_version SET version = 66 WHERE version = 65;
