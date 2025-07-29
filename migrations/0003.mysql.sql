CREATE TABLE IF NOT EXISTS `sessions` (
  `session_id` varchar(128) NOT NULL,
  `users_idusers` int NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`session_id`),
  KEY `sessions_user_idx` (`users_idusers`)
);

-- Record upgrade to schema version 3
UPDATE schema_version SET version = 3 WHERE version = 2;
