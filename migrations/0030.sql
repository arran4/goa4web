ALTER TABLE blogs
    MODIFY COLUMN forumthread_id int(10) DEFAULT NULL;

-- Ensure each forum topic only has one restrictions row
ALTER TABLE topicrestrictions
    ADD UNIQUE INDEX topicrestrictions_forumtopic_idx (forumtopic_idforumtopic);
ALTER TABLE topicrestrictions
    ADD PRIMARY KEY (forumtopic_idforumtopic);

-- Record upgrade to schema version 30
UPDATE schema_version SET version = 30 WHERE version = 29;
