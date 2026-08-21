-- +goose Up
ALTER TABLE forumthread ADD COLUMN reply_to_comment_id INTEGER DEFAULT NULL;
ALTER TABLE forumthread ADD COLUMN reply_to_thread_id INTEGER DEFAULT NULL;
CREATE INDEX IF NOT EXISTS forumthread_reply_to_thread_id ON forumthread(reply_to_thread_id, reply_to_comment_id);

UPDATE schema_version SET version = 95;
