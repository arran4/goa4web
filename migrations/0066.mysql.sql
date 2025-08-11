ALTER TABLE forumcategory
    CHANGE COLUMN idforumcategory id int(10) NOT NULL AUTO_INCREMENT,
    CHANGE COLUMN forumcategory_idforumcategory parent_category_id int(10) NOT NULL DEFAULT 0,
    CHANGE COLUMN language_idlanguage language_id int(10) DEFAULT NULL;

ALTER TABLE forumthread
    CHANGE COLUMN idforumthread id int(10) NOT NULL AUTO_INCREMENT,
    CHANGE COLUMN forumtopic_idforumtopic topic_id int(10) NOT NULL DEFAULT 0,
    CHANGE COLUMN firstpost first_comment_id int(10) NOT NULL DEFAULT 0,
    CHANGE COLUMN lastposter last_author_id int(10) NOT NULL DEFAULT 0;

ALTER TABLE forumtopic
    CHANGE COLUMN idforumtopic id int(10) NOT NULL AUTO_INCREMENT,
    CHANGE COLUMN forumcategory_idforumcategory category_id int(10) NOT NULL DEFAULT 0,
    CHANGE COLUMN language_idlanguage language_id int(10) DEFAULT NULL,
    CHANGE COLUMN lastposter last_author_id int(10) NOT NULL DEFAULT 0;

-- Update schema version
UPDATE schema_version SET version = 66 WHERE version = 65;
