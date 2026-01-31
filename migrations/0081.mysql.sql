ALTER TABLE external_links ADD COLUMN card_duration TINYTEXT;
ALTER TABLE external_links ADD COLUMN card_upload_date TINYTEXT;
ALTER TABLE external_links ADD COLUMN card_author TINYTEXT;

UPDATE schema_version SET version = 81 WHERE version = 80;
