-- Rename writing column and approval table
ALTER TABLE writing CHANGE COLUMN writting writing LONGTEXT;
ALTER TABLE deactivated_writings CHANGE COLUMN writting writing LONGTEXT;
ALTER TABLE writtingApprovedUsers RENAME TO writingApprovedUsers;

-- Record upgrade to schema version 18
UPDATE schema_version SET version = 18 WHERE version = 17;
