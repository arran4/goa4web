-- +goose Up
RENAME TABLE writingCategory TO writing_category;

-- Record upgrade to schema version 22
