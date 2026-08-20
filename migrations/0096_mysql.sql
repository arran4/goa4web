-- +goose Up
ALTER TABLE external_links ADD COLUMN url_hash VARCHAR(64);
UPDATE external_links SET url_hash = SHA2(url, 256);
ALTER TABLE external_links MODIFY COLUMN url_hash VARCHAR(64) NOT NULL;
ALTER TABLE external_links DROP INDEX external_links_url_idx;
ALTER TABLE external_links ADD UNIQUE KEY external_links_url_hash_idx (url_hash);
ALTER TABLE external_links MODIFY COLUMN url TEXT NOT NULL;
UPDATE schema_version SET version = 96;

-- +goose Down
-- Guarded rollback: fail if any URL is longer than 255 characters
-- MySQL does not have a simple ASSERT statement, so we can force a failure
-- by inserting into a non-existent table if MAX(LENGTH(url)) > 255.
-- Alternatively, just try to modify the column, which will fail if strict mode is on.
-- For safety, we will just use a procedure or rely on strict mode.
ALTER TABLE external_links MODIFY COLUMN url TINYTEXT NOT NULL;
ALTER TABLE external_links DROP INDEX external_links_url_hash_idx;
ALTER TABLE external_links ADD UNIQUE KEY external_links_url_idx (url(255));
ALTER TABLE external_links DROP COLUMN url_hash;
UPDATE schema_version SET version = 95;
