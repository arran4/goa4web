-- Store template overrides
CREATE TABLE IF NOT EXISTS `template_overrides` (
    `name` varchar(128) NOT NULL,
    `body` text NOT NULL,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`name`)
);

-- Record upgrade to schema version 7
UPDATE schema_version SET version = 7 WHERE version = 6;
