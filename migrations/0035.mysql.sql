-- Rename user_roles primary key column
ALTER TABLE user_roles CHANGE COLUMN idpermissions iduser_roles INT NOT NULL AUTO_INCREMENT;

-- Update schema version
UPDATE schema_version SET version = 35 WHERE version = 34;
