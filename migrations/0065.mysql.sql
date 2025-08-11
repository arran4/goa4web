RENAME TABLE forumthread_public_labels TO content_public_labels;
ALTER TABLE content_public_labels
    CHANGE COLUMN forumthread_idforumthread item_id INT NOT NULL,
    ADD COLUMN item VARCHAR(64) NOT NULL DEFAULT 'thread' AFTER id,
    DROP INDEX forumthread_public_labels_uq,
    ADD UNIQUE INDEX content_public_labels_uq (item, item_id, label(255));
ALTER TABLE content_public_labels MODIFY item VARCHAR(64) NOT NULL;

RENAME TABLE forumthread_private_labels TO content_private_labels;
ALTER TABLE content_private_labels
    CHANGE COLUMN forumthread_idforumthread item_id INT NOT NULL,
    CHANGE COLUMN users_idusers user_id INT NOT NULL,
    ADD COLUMN item VARCHAR(64) NOT NULL DEFAULT 'thread' AFTER id,
    DROP INDEX forumthread_private_labels_uq,
    ADD UNIQUE INDEX content_private_labels_uq (item, item_id, user_id, label(255));
ALTER TABLE content_private_labels MODIFY item VARCHAR(64) NOT NULL;

-- Update schema version
UPDATE schema_version SET version = 65 WHERE version = 64;
