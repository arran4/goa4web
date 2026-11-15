//go:build !sqlite && !sqlite3

package db

import "testing"

func TestNotificationReadStateLinearizabilityMySQL(t *testing.T) {
	database := openReplyThreadTestDatabase(t)
	execReplySQL(t, database,
		`CREATE TABLE content_read_markers (
			id int NOT NULL AUTO_INCREMENT,
			item varchar(32) NOT NULL,
			item_id int NOT NULL,
			user_id int NOT NULL,
			last_comment_id int NOT NULL,
			last_activity_position bigint NOT NULL DEFAULT '0',
			PRIMARY KEY (id),
			UNIQUE KEY content_read_markers_uq (item, item_id, user_id)
		)`,
		`CREATE TABLE notifications (
			id int NOT NULL AUTO_INCREMENT,
			users_idusers int NOT NULL,
			link varchar(2000) DEFAULT NULL,
			message text,
			activity_key varchar(255) DEFAULT NULL,
			priority tinyint unsigned NOT NULL DEFAULT '0',
			resource_type varchar(64) DEFAULT NULL,
			resource_id int DEFAULT NULL,
			activity_position bigint DEFAULT NULL,
			created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
			read_at timestamp NULL DEFAULT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY notifications_recipient_activity_uq (users_idusers, activity_key),
			KEY notifications_recipient_resource_activity_idx (users_idusers, resource_type, resource_id, activity_position)
		)`,
	)

	queries := NewForDriver(database, "mysql")
	runNotificationReadStateLinearizabilitySuite(t, queries)
}
