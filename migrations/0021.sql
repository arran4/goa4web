ALTER TABLE topicrestrictions
    DROP INDEX threadrestrictions_FKIndex1,
    ADD PRIMARY KEY (forumtopic_idforumtopic);

-- Record upgrade to schema version 21
UPDATE schema_version SET version = 21 WHERE version = 20;
