-- Create API Keys table
CREATE TABLE IF NOT EXISTS `api_keys` (
  `id` int NOT NULL AUTO_INCREMENT,
  `users_idusers` int NOT NULL,
  `api_key` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `scopes` text NOT NULL,
  `expires_at` datetime DEFAULT NULL,
  `last_used_at` datetime DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `revoked_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `api_keys_key_idx` (`api_key`)
);

-- Set the schema version to the latest migration.
UPDATE schema_version SET version = 89 WHERE version = 88;
