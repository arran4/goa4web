-- Rename topic and user level tables and columns
RENAME TABLE topicrestrictions TO topic_permissions;
RENAME TABLE userstopiclevel TO user_topic_permissions;
RENAME TABLE writing_approved_users TO writing_user_permissions;

ALTER TABLE user_topic_permissions CHANGE COLUMN level role_id INT NULL;
ALTER TABLE topic_permissions
  CHANGE COLUMN viewlevel view_role_id INT NULL,
  CHANGE COLUMN replylevel reply_role_id INT NULL,
  CHANGE COLUMN newthreadlevel newthread_role_id INT NULL,
  CHANGE COLUMN seelevel see_role_id INT NULL,
  CHANGE COLUMN invitelevel invite_role_id INT NULL,
  CHANGE COLUMN readlevel read_role_id INT NULL,
  CHANGE COLUMN modlevel mod_role_id INT NULL,
  CHANGE COLUMN adminlevel admin_role_id INT NULL;

ALTER TABLE writing_user_permissions
  CHANGE COLUMN readdoc can_read TINYINT(1) NULL,
  CHANGE COLUMN editdoc can_edit TINYINT(1) NULL;

-- Update schema version
UPDATE schema_version SET version = 36 WHERE version = 35;
