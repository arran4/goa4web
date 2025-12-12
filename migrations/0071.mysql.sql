ALTER TABLE imageboard ADD COLUMN deleted_at DATETIME DEFAULT NULL;

INSERT INTO schema_version (version) VALUES (71) ON DUPLICATE KEY UPDATE version = VALUES(version);
