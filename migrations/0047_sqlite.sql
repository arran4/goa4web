-- +goose Up
ALTER TABLE audit_log ADD COLUMN path TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN details TEXT;
ALTER TABLE audit_log ADD COLUMN data TEXT;
UPDATE schema_version SET version = 47;