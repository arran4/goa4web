-- Rename misnamed writtingApprovedUsers table
RENAME TABLE writtingApprovedUsers TO writingApprovedUsers;

-- Record upgrade to schema version 18
UPDATE schema_version SET version = 18 WHERE version = 17;

