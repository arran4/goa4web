-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS users_username_idx ON users (username);
CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx ON users (email);
CREATE UNIQUE INDEX IF NOT EXISTS topicrestrictions_topic_idx ON topicrestrictions (forumtopic_idforumtopic);
UPDATE schema_version SET version = 21;