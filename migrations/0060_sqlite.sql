-- +goose Up
UPDATE blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE faq SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE linker_queue SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE preferences SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE site_news SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE writing SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE deactivated_comments SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE deactivated_writings SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE deactivated_blogs SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE deactivated_linker SET language_idlanguage = NULL WHERE language_idlanguage = 0;
UPDATE blogs SET forumthread_id = NULL WHERE forumthread_id = 0;
UPDATE schema_version SET version = 60;