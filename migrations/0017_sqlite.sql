-- +goose Up
ALTER TABLE siteNews RENAME COLUMN occured TO occurred;
UPDATE schema_version SET version = 17;