-- +goose Up
ALTER TABLE forumcategory ADD COLUMN language_idlanguage INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS forumcategory_FKIndex2 ON forumcategory (language_idlanguage);

ALTER TABLE forumtopic ADD COLUMN language_idlanguage INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS forumtopic_FKIndex3 ON forumtopic (language_idlanguage);
UPDATE schema_version SET version = 54;