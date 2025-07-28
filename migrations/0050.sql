-- Add unique index for searchwordlist words
ALTER TABLE searchwordlist
    ADD UNIQUE KEY searchwordlist_word_idx (word(255));

-- Update schema version
UPDATE schema_version SET version = 50 WHERE version = 49;
