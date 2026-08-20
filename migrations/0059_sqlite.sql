-- +goose Up
UPDATE writing_category SET writing_category_id = NULL WHERE writing_category_id = 0;
UPDATE schema_version SET version = 59;