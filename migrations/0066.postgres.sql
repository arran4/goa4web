ALTER TABLE language RENAME COLUMN idlanguage TO id;
ALTER TABLE blogs RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE comments RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE faq RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE forumcategory RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE forumtopic RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE linker RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE linker_queue RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE preferences RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE site_news RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE writing RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE deactivated_comments RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE deactivated_writings RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE deactivated_blogs RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE deactivated_linker RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE user_language RENAME COLUMN language_idlanguage TO language_id;

-- Update schema version
UPDATE schema_version SET version = 66 WHERE version = 65;
