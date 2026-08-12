-- +goose Up
ALTER TABLE `forumthread` ADD COLUMN `reply_to_comment_id` int(10) DEFAULT NULL;
ALTER TABLE `forumthread` ADD COLUMN `reply_to_thread_id` int(10) DEFAULT NULL;
