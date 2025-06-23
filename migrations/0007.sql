-- Add audit_log table for tracking admin actions
CREATE TABLE IF NOT EXISTS audit_log (
    id INT NOT NULL AUTO_INCREMENT,
    users_idusers INT NOT NULL,
    action TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY audit_log_user_idx (users_idusers)
);

-- Record upgrade to schema version 7
UPDATE schema_version SET version = 7 WHERE version = 6;
