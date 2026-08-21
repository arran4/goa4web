-- +goose Up
ALTER TABLE topicrestrictions RENAME TO topic_permissions;
ALTER TABLE userstopiclevel RENAME TO user_topic_permissions;
ALTER TABLE writing_approved_users RENAME TO writing_user_permissions;

ALTER TABLE user_topic_permissions RENAME COLUMN level TO role_id;

ALTER TABLE topic_permissions RENAME COLUMN viewlevel TO view_role_id;
ALTER TABLE topic_permissions RENAME COLUMN replylevel TO reply_role_id;
ALTER TABLE topic_permissions RENAME COLUMN newthreadlevel TO newthread_role_id;
ALTER TABLE topic_permissions RENAME COLUMN seelevel TO see_role_id;
ALTER TABLE topic_permissions RENAME COLUMN invitelevel TO invite_role_id;
ALTER TABLE topic_permissions RENAME COLUMN readlevel TO read_role_id;
ALTER TABLE topic_permissions RENAME COLUMN modlevel TO mod_role_id;
ALTER TABLE topic_permissions RENAME COLUMN adminlevel TO admin_role_id;

ALTER TABLE writing_user_permissions RENAME COLUMN readdoc TO can_read;
ALTER TABLE writing_user_permissions RENAME COLUMN editdoc TO can_edit;

UPDATE schema_version SET version = 36;