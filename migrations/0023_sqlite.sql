-- +goose Up
ALTER TABLE writing RENAME COLUMN writingCategory_idwritingCategory TO writing_category_id;
ALTER TABLE writing_category RENAME COLUMN writingCategory_idwritingCategory TO writing_category_id;
UPDATE schema_version SET version = 23;