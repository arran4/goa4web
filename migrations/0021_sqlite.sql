-- +goose Up
ALTER TABLE users
    ADD UNIQUE INDEX users_username_idx (username(255)),
    ADD UNIQUE INDEX users_email_idx (email(255));

ALTER TABLE topicrestrictions
    DROP INDEX threadrestrictions_FKIndex1,
    ADD PRIMARY KEY (forumtopic_idforumtopic);