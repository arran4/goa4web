-- +goose Up
ALTER TABLE forumcategory ADD COLUMN language_idlanguage INT NOT NULL DEFAULT 0;

CREATE INDEX forumcategory_FKIndex2 ON forumcategory (language_idlanguage);

ALTER TABLE forumtopic ADD COLUMN language_idlanguage INT NOT NULL DEFAULT 0;

CREATE INDEX forumtopic_FKIndex3 ON forumtopic (language_idlanguage);