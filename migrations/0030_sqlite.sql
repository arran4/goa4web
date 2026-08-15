-- +goose Up
-- ALTER TABLE permissions
    CHANGE COLUMN level role tinyblob DEFAULT NULL;

-- ALTER TABLE blogs
    MODIFY COLUMN forumthread_id int(10) DEFAULT NULL;

ALTER TABLE topicrestrictions
    ADD UNIQUE INDEX topicrestrictions_forumtopic_idx (forumtopic_idforumtopic);