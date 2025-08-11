ALTER TABLE forumthread_public_labels RENAME TO content_public_labels;
ALTER TABLE content_public_labels RENAME COLUMN forumthread_idforumthread TO item_id;
ALTER TABLE content_public_labels ADD COLUMN item TEXT NOT NULL DEFAULT 'thread';
ALTER TABLE content_public_labels DROP CONSTRAINT forumthread_public_labels_uq;
ALTER TABLE content_public_labels ADD CONSTRAINT content_public_labels_uq UNIQUE (item, item_id, label);
ALTER TABLE content_public_labels ALTER COLUMN item DROP DEFAULT;

ALTER TABLE forumthread_private_labels RENAME TO content_private_labels;
ALTER TABLE content_private_labels RENAME COLUMN forumthread_idforumthread TO item_id;
ALTER TABLE content_private_labels RENAME COLUMN users_idusers TO user_id;
ALTER TABLE content_private_labels ADD COLUMN item TEXT NOT NULL DEFAULT 'thread';
ALTER TABLE content_private_labels DROP CONSTRAINT forumthread_private_labels_uq;
ALTER TABLE content_private_labels ADD CONSTRAINT content_private_labels_uq UNIQUE (item, item_id, user_id, label);
ALTER TABLE content_private_labels ALTER COLUMN item DROP DEFAULT;

-- Update schema version
UPDATE schema_version SET version = 65 WHERE version = 64;
