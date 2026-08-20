-- +goose Up
ALTER TABLE preferences ADD COLUMN page_size INT NOT NULL DEFAULT 15;
ALTER TABLE linkerCategory ADD COLUMN position INT NOT NULL DEFAULT 0;
ALTER TABLE linkerCategory ADD COLUMN sortorder INT NOT NULL DEFAULT 0;
DROP TABLE IF EXISTS sidTable;
UPDATE schema_version SET version = 2;