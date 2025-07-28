ALTER TABLE searchwordlist
    ADD UNIQUE KEY searchwordlist_word_idx (word(255));

-- Create FAQ revision history table
CREATE TABLE IF NOT EXISTS faq_revisions (
    id INT NOT NULL AUTO_INCREMENT,
    faq_id INT NOT NULL,
    users_idusers INT NOT NULL,
    question MEDIUMTEXT,
    answer MEDIUMTEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(id),
    KEY faq_revisions_faq_idx (faq_id)
);

-- Record upgrade to schema version 50
UPDATE schema_version SET version = 50 WHERE version = 49;
