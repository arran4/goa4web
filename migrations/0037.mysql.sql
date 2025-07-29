-- Add grants table for new permission model
CREATE TABLE grants (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  created_at DATETIME NULL,
  updated_at DATETIME NULL,
  user_id INT NULL,
  role_id INT NULL,
  section VARCHAR(64) NOT NULL,
  item VARCHAR(64) NULL,
  rule_type VARCHAR(32) NOT NULL,
  item_id INT NULL,
  item_rule VARCHAR(64) NULL,
  action VARCHAR(64) NOT NULL,
  extra VARCHAR(64) NULL,
  active TINYINT(1) NOT NULL DEFAULT 1
);

-- Update schema version
UPDATE schema_version SET version = 37 WHERE version = 36;
