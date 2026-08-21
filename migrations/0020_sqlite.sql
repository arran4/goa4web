-- +goose Up
ALTER TABLE uploaded_images ADD COLUMN width INT DEFAULT NULL;
ALTER TABLE uploaded_images ADD COLUMN height INT DEFAULT NULL;
ALTER TABLE commentsSearch RENAME COLUMN comments_idcomments TO comment_id;
ALTER TABLE imagepostSearch RENAME COLUMN imagepost_idimagepost TO image_post_id;
ALTER TABLE writingSearch RENAME COLUMN writing_idwriting TO writing_id;
ALTER TABLE writingApprovedUsers RENAME COLUMN writing_idwriting TO writing_id;
ALTER TABLE linkerSearch RENAME COLUMN linker_idlinker TO linker_id;
ALTER TABLE searchwordlist_has_linker RENAME COLUMN linker_idlinker TO linker_id;
UPDATE schema_version SET version = 20;