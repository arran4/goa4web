-- +goose Up
ALTER TABLE writtingApprovedUsers RENAME TO writingApprovedUsers;
ALTER TABLE writing RENAME COLUMN writting TO writing;
ALTER TABLE deactivated_writings RENAME COLUMN writting TO writing;
ALTER TABLE userlang RENAME TO user_language;
UPDATE schema_version SET version = 18;