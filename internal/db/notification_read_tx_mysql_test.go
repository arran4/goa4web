//go:build !sqlite && !sqlite3

package db

import (
	"context"
	"testing"
)

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

	t.Run("keep thread unread through a partial append acknowledgment", func(t *testing.T) {
		execReplySQL(t, database,
			`CREATE TABLE content_private_labels (
				item varchar(64) NOT NULL, item_id int NOT NULL, user_id int NOT NULL,
				label varchar(32) NOT NULL, invert boolean NOT NULL,
				UNIQUE (item, item_id, user_id, label)
			)`,
			`CREATE TABLE comments (idcomments int PRIMARY KEY, forumthread_id int, language_id int, deleted_at datetime)`,
			`CREATE TABLE comment_appends (id int PRIMARY KEY, comment_id int)`,
			`CREATE TABLE user_language (users_idusers int, language_id int)`,
			`INSERT INTO comments (idcomments, forumthread_id, language_id) VALUES (400, 500, 1), (401, 500, 2)`,
			`INSERT INTO comment_appends (id, comment_id) VALUES (20, 400), (21, 400)`,
			`INSERT INTO user_language (users_idusers, language_id) VALUES (600, 1)`,
		)
		ctx := context.Background()
		acknowledge := func(appendID int64, wantRead bool) {
			t.Helper()
			if err := queries.(NotificationReadStateCoordinator).AcknowledgeThreadReadActivity(ctx, AcknowledgeThreadReadActivityParams{
				RecipientID: 600, ThreadID: 500, CommentID: 400,
				ActivityPosition: int64(400)<<32 | appendID,
				UpdateLabels:     true,
			}); err != nil {
				t.Fatalf("acknowledge append %d: %v", appendID, err)
			}
			labels, err := queries.ListContentPrivateLabels(ctx, ListContentPrivateLabelsParams{
				Item: "thread", ItemID: 500, UserID: 600,
			})
			if err != nil {
				t.Fatalf("list read labels: %v", err)
			}
			read := false
			for _, label := range labels {
				if label.Label == "unread" && label.Invert {
					read = true
				}
			}
			if read != wantRead {
				t.Fatalf("after append %d, read = %t, want %t", appendID, read, wantRead)
			}
		}
		acknowledge(20, false)
		acknowledge(21, true)
		acknowledge(20, true)
	})
}
