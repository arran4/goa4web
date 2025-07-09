ALTER TABLE uploaded_images
  DROP COLUMN thumbnail,
  ADD COLUMN width INT DEFAULT NULL,
  ADD COLUMN height INT DEFAULT NULL;

ALTER TABLE commentsSearch CHANGE COLUMN comments_idcomments comment_id int(10) NOT NULL DEFAULT 0;

-- Rename imagepost search column to snake_case
ALTER TABLE imagepostSearch CHANGE COLUMN imagepost_idimagepost image_post_id int(10) NOT NULL DEFAULT 0;

-- Rename writing search foreign keys to snake_case
ALTER TABLE writingSearch CHANGE COLUMN writing_idwriting writing_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE writingApprovedUsers CHANGE COLUMN writing_idwriting writing_id int(10) NOT NULL DEFAULT 0;

-- Rename linker search columns to snake_case
ALTER TABLE linkerSearch CHANGE COLUMN linker_idlinker linker_id int(10) NOT NULL DEFAULT 0;
ALTER TABLE searchwordlist_has_linker CHANGE COLUMN linker_idlinker linker_id int(10) NOT NULL DEFAULT 0;

-- Record upgrade to schema version 20
UPDATE schema_version SET version = 20 WHERE version = 19;
