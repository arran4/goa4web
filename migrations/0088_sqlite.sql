-- +goose Up
ALTER TABLE image_cache_entries
  ADD COLUMN IF NOT EXISTS uploaded_image_id INT DEFAULT NULL,
  ADD KEY IF NOT EXISTS image_cache_entries_uploaded_image_idx (uploaded_image_id);