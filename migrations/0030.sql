ALTER TABLE topicrestrictions
    ADD PRIMARY KEY (forumtopic_idforumtopic);

-- Record upgrade to schema version 30
UPDATE schema_version SET version = 30 WHERE version = 29;
