-- Drop legacy writing_user_permissions table
DROP TABLE IF EXISTS writing_user_permissions;

-- Introduce admin user comments and rejection role
CREATE TABLE admin_user_comments (
  id INT NOT NULL AUTO_INCREMENT,
  users_idusers INT NOT NULL,
  comment TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY admin_user_comments_user_idx (users_idusers)
);

INSERT INTO roles (name) VALUES ('rejected') ON DUPLICATE KEY UPDATE name=name;

-- Update schema version
UPDATE schema_version SET version = 40 WHERE version = 39;
