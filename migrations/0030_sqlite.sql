-- +goose Up
ALTER TABLE permissions RENAME COLUMN level TO role;
CREATE UNIQUE INDEX IF NOT EXISTS topicrestrictions_forumtopic_idx ON topicrestrictions (forumtopic_idforumtopic);
UPDATE schema_version SET version = 30;