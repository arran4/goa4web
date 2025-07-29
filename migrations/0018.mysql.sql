-- Rename misnamed writtingApprovedUsers table
RENAME TABLE writtingApprovedUsers TO writingApprovedUsers;

-- Rename writing column and approval table
ALTER TABLE writing CHANGE COLUMN writting writing LONGTEXT;
ALTER TABLE deactivated_writings CHANGE COLUMN writting writing LONGTEXT;
-- Rename userlang table to user_language
ALTER TABLE userlang RENAME TO user_language;

-- Record upgrade to schema version 18
UPDATE schema_version SET version = 18 WHERE version = 17;
