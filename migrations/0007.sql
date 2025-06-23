-- Add banned_ips table with expiry and cancellation tracking
CREATE TABLE IF NOT EXISTS `banned_ips` (
    `id` int NOT NULL AUTO_INCREMENT,
    `ip_net` varchar(50) NOT NULL,
    `reason` text,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `expires_at` datetime DEFAULT NULL,
    `canceled_at` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `banned_ips_ip_idx` (`ip_net`)
);

-- Record upgrade to schema version 7
UPDATE schema_version SET version = 7 WHERE version = 6;
