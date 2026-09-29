-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pending_actions (
    id VARCHAR(128) NOT NULL PRIMARY KEY,
    uid INT NOT NULL,
    browser_id VARCHAR(128) NOT NULL,
    action_type VARCHAR(128) NOT NULL,
    form_data MEDIUMTEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    consumed_at TIMESTAMP NULL,
    FOREIGN KEY (uid) REFERENCES users(idusers) ON DELETE CASCADE
);
CREATE INDEX idx_pending_actions_expires_at ON pending_actions (expires_at);
CREATE INDEX idx_pending_actions_uid ON pending_actions (uid);
-- +goose StatementEnd
