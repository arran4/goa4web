-- Upgrade from schema version 1 (v0.0.1) to 2 (v0.0.2)
-- Adds new notification and email queue tables and updates existing structures.

-- Add page_size column to preferences if it doesn't exist
ALTER TABLE preferences
    ADD COLUMN IF NOT EXISTS page_size INT NOT NULL DEFAULT 15;

-- Add position and sortorder columns to linkerCategory if they don't exist
ALTER TABLE linkerCategory
    ADD COLUMN IF NOT EXISTS position INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sortorder INT NOT NULL DEFAULT 0;

-- Ensure blogs.written has a timestamp default
ALTER TABLE blogs
    MODIFY written DATETIME NOT NULL DEFAULT NOW();

-- Drop obsolete sidTable
DROP TABLE IF EXISTS sidTable;

-- Core new tables for v0.0.2
CREATE TABLE IF NOT EXISTS schema_version (
    version INT NOT NULL
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id INT NOT NULL AUTO_INCREMENT,
    users_idusers INT NOT NULL,
    item_type VARCHAR(32) NOT NULL,
    target_id INT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS pending_emails (
    id INT NOT NULL AUTO_INCREMENT,
    to_email TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at DATETIME DEFAULT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS notifications (
    id INT NOT NULL AUTO_INCREMENT,
    users_idusers INT NOT NULL,
    link TEXT,
    message TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    read_at DATETIME DEFAULT NULL,
    PRIMARY KEY (id)
);

-- Update schema version
INSERT INTO schema_version (version) VALUES (2)
    ON DUPLICATE KEY UPDATE version = VALUES(version);
