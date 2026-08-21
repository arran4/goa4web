-- +goose Up
ALTER TABLE writingCategory RENAME TO writing_category;
UPDATE schema_version SET version = 22;