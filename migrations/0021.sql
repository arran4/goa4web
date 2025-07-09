-- Add unique indexes to username and email
ALTER TABLE users
    ADD UNIQUE INDEX users_username_idx (username(255)),
    ADD UNIQUE INDEX users_email_idx (email(255));

-- Record upgrade to schema version 21
UPDATE schema_version SET version = 21 WHERE version = 20;

ALTER TABLE topicrestrictions
    DROP INDEX threadrestrictions_FKIndex1,
    ADD PRIMARY KEY (forumtopic_idforumtopic);

-- Record upgrade to schema version 21
UPDATE schema_version SET version = 21 WHERE version = 20;
