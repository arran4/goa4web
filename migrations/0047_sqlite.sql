-- +goose Up
ALTER TABLE audit_log
    ADD COLUMN path TEXT NOT NULL DEFAULT '' AFTER action,
    ADD COLUMN details TEXT AFTER path,
    ADD COLUMN data TEXT AFTER details;