-- Rename writing search foreign keys to snake_case
ALTER TABLE writingSearch CHANGE COLUMN writing_idwriting writing_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE writingApprovedUsers CHANGE COLUMN writing_idwriting writing_id int(10) NOT NULL DEFAULT 0;

-- Rename linker search columns to snake_case
ALTER TABLE linkerSearch CHANGE COLUMN linker_idlinker linker_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE searchwordlist_has_linker CHANGE COLUMN linker_idlinker linker_id int(10) NOT NULL DEFAULT 0;

-- Record upgrade to schema version 20
UPDATE schema_version SET version = 20 WHERE version = 19;
