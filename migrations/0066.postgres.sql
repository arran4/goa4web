ALTER TABLE forumcategory RENAME COLUMN idforumcategory TO id;
ALTER TABLE forumcategory RENAME COLUMN forumcategory_idforumcategory TO parent_category_id;
ALTER TABLE forumcategory RENAME COLUMN language_idlanguage TO language_id;

ALTER TABLE forumthread RENAME COLUMN idforumthread TO id;
ALTER TABLE forumthread RENAME COLUMN forumtopic_idforumtopic TO topic_id;
ALTER TABLE forumthread RENAME COLUMN firstpost TO first_comment_id;
ALTER TABLE forumthread RENAME COLUMN lastposter TO last_author_id;

ALTER TABLE forumtopic RENAME COLUMN idforumtopic TO id;
ALTER TABLE forumtopic RENAME COLUMN forumcategory_idforumcategory TO category_id;
ALTER TABLE forumtopic RENAME COLUMN language_idlanguage TO language_id;
ALTER TABLE forumtopic RENAME COLUMN lastposter TO last_author_id;

-- Update schema version
UPDATE schema_version SET version = 66 WHERE version = 65;
