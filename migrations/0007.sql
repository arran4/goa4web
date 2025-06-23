-- Store template overrides
CREATE TABLE IF NOT EXISTS `template_overrides` (
    `name` varchar(128) NOT NULL,
    `body` text NOT NULL,
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`name`)
);

-- Add audit_log table for tracking admin actions
CREATE TABLE IF NOT EXISTS audit_log (
    id INT NOT NULL AUTO_INCREMENT,
    users_idusers INT NOT NULL,
    action TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY audit_log_user_idx (users_idusers)
);

-- Add passwd_algorithm column to track the password hashing scheme
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS passwd_algorithm TINYTEXT DEFAULT NULL;

-- Migrate existing users to the legacy md5 algorithm
UPDATE users SET passwd_algorithm = 'md5' WHERE passwd_algorithm IS NULL;

-- Record upgrade to schema version 7
UPDATE schema_version SET version = 7 WHERE version = 6;
