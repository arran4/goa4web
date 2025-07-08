-- Rename imagepost search column to snake_case
ALTER TABLE imagepostSearch CHANGE COLUMN imagepost_idimagepost image_post_id int(10) NOT NULL DEFAULT 0;

-- Record upgrade to schema version 20
UPDATE schema_version SET version = 20 WHERE version = 19;
