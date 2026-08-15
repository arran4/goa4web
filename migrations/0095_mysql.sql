-- +goose Up
ALTER TABLE forumthread
  ADD COLUMN reply_to_comment_id int(10) DEFAULT NULL,
  ADD COLUMN reply_to_thread_id int(10) DEFAULT NULL;

UPDATE schema_version SET version = 95;
