ALTER TABLE forumthread_public_labels RENAME TO content_public_labels;
ALTER TABLE content_public_labels RENAME COLUMN forumthread_idforumthread TO item_id;
ALTER TABLE content_public_labels ADD COLUMN item TEXT NOT NULL DEFAULT 'thread';
CREATE UNIQUE INDEX content_public_labels_uq ON content_public_labels (item, item_id, label);
DROP INDEX forumthread_public_labels_uq;

ALTER TABLE forumthread_private_labels RENAME TO content_private_labels;
ALTER TABLE content_private_labels RENAME COLUMN forumthread_idforumthread TO item_id;
ALTER TABLE content_private_labels RENAME COLUMN users_idusers TO user_id;
ALTER TABLE content_private_labels ADD COLUMN item TEXT NOT NULL DEFAULT 'thread';
CREATE UNIQUE INDEX content_private_labels_uq ON content_private_labels (item, item_id, user_id, label);
DROP INDEX forumthread_private_labels_uq;

-- Update schema version
UPDATE schema_version SET version = 65 WHERE version = 64;
