-- +goose Up
ALTER TABLE user_roles RENAME COLUMN idpermissions TO iduser_roles;
UPDATE schema_version SET version = 35;