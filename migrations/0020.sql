-- Rename linker search columns to snake_case
ALTER TABLE linkerSearch CHANGE COLUMN linker_idlinker linker_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE searchwordlist_has_linker CHANGE COLUMN linker_idlinker linker_id int(10) NOT NULL DEFAULT 0;

-- Record upgrade to schema version 20
UPDATE schema_version SET version = 20 WHERE version = 19;
