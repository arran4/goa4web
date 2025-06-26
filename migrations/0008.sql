-- Add tables to retain deactivated data
CREATE TABLE IF NOT EXISTS deactivated_users LIKE users;
CREATE TABLE IF NOT EXISTS deactivated_comments LIKE comments;

-- Record upgrade to schema version 8
UPDATE schema_version SET version = 8 WHERE version = 7;
