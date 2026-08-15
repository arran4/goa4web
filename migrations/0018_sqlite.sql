-- +goose Up
RENAME TABLE writtingApprovedUsers TO writingApprovedUsers;

-- ALTER TABLE writing CHANGE COLUMN writting writing LONGTEXT;

-- ALTER TABLE deactivated_writings CHANGE COLUMN writting writing LONGTEXT;

ALTER TABLE userlang RENAME TO user_language;