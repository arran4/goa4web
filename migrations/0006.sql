-- Add site_announcements table
CREATE TABLE IF NOT EXISTS `site_announcements` (
    `id` int NOT NULL AUTO_INCREMENT,
    `site_news_id` int NOT NULL,
    `active` tinyint(1) NOT NULL DEFAULT 1,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `site_announcements_news_idx` (`site_news_id`)
);
-- Track failed login attempts.
CREATE TABLE IF NOT EXISTS `login_attempts` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `ip_address` varchar(45) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
);

-- Record upgrade to schema version 6
UPDATE schema_version SET version = 6 WHERE version = 5;
