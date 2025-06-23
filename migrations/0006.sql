-- Add site_announcements table
CREATE TABLE IF NOT EXISTS `site_announcements` (
    `id` int NOT NULL AUTO_INCREMENT,
    `site_news_id` int NOT NULL,
    `active` tinyint(1) NOT NULL DEFAULT 1,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `site_announcements_news_idx` (`site_news_id`)
);
