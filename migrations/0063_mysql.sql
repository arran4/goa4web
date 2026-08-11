-- +goose Up
ALTER TABLE comments ADD COLUMN timezone TINYTEXT DEFAULT NULL;
ALTER TABLE deactivated_comments ADD COLUMN timezone TINYTEXT DEFAULT NULL;


