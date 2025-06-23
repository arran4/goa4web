-- Add banned_ips table
CREATE TABLE IF NOT EXISTS `banned_ips` (
    `id` int NOT NULL AUTO_INCREMENT,
    `ip_address` varchar(45) NOT NULL,
    `reason` text,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `banned_ips_ip_idx` (`ip_address`)
);

-- Record upgrade to schema version 7
UPDATE schema_version SET version = 7 WHERE version = 6;
