-- +goose Up
UPDATE grants SET section = 'writing' WHERE section = 'writings';
UPDATE schema_version SET version = 53;