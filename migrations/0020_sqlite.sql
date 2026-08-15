-- +goose Up
-- ALTER TABLE uploaded_images
  DROP COLUMN thumbnail,
  ADD COLUMN width INT DEFAULT NULL,
  ADD COLUMN height INT DEFAULT NULL;

-- ALTER TABLE commentsSearch CHANGE COLUMN comments_idcomments comment_id int(10) NOT NULL DEFAULT 0;

-- ALTER TABLE imagepostSearch CHANGE COLUMN imagepost_idimagepost image_post_id int(10) NOT NULL DEFAULT 0;

-- ALTER TABLE writingSearch CHANGE COLUMN writing_idwriting writing_id int(10) NOT NULL DEFAULT 0;

-- ALTER TABLE writingApprovedUsers CHANGE COLUMN writing_idwriting writing_id int(10) NOT NULL DEFAULT 0;

-- ALTER TABLE linkerSearch CHANGE COLUMN linker_idlinker linker_id int(10) NOT NULL DEFAULT 0;

-- ALTER TABLE searchwordlist_has_linker CHANGE COLUMN linker_idlinker linker_id int(10) NOT NULL DEFAULT 0;