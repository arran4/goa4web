ALTER TABLE pending_passwords ADD COLUMN expires_at TIMESTAMP NULL;
ALTER TABLE pending_passwords MODIFY COLUMN passwd VARCHAR(255) NULL;
ALTER TABLE pending_passwords MODIFY COLUMN passwd_algorithm VARCHAR(255) NULL;
