ALTER TABLE faq ADD COLUMN priority INT NOT NULL DEFAULT 0;
CREATE INDEX faq_priority_idx ON faq (priority);
