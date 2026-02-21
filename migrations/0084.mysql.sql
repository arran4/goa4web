ALTER TABLE topicrestrictions ADD PRIMARY KEY (forumtopic_idforumtopic);

UPDATE schema_version SET version = 84 WHERE version = 83;
