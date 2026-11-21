//go:build !sqlite && !sqlite3

package db

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/database"
	"github.com/stretchr/testify/require"
)

func TestMySQLRecordsAndSummarizesAdminJobFailures(t *testing.T) {
	conn := openReplyThreadTestDatabase(t)
	ctx := t.Context()
	// notificationTable identifies the canonical table used by this query contract.
	const notificationTable = "CREATE TABLE IF NOT EXISTS `notifications` ("
	_, rest, found := strings.Cut(string(database.SchemaMySQL), notificationTable)
	require.True(t, found, "canonical notifications table")
	definition, _, found := strings.Cut(rest, ";")
	require.True(t, found, "complete notifications table definition")
	execReplySQL(t, conn, notificationTable+definition)
	q := NewForDriver(conn, "mysql")
	key := sql.NullString{String: "failure-group", Valid: true}
	now := time.Now().UTC().Truncate(time.Second)
	record := func(occurredAt time.Time) {
		t.Helper()
		require.NoError(t, q.(JobTransactor).WithJobTransaction(ctx, func(tx Querier) error {
			if err := tx.SystemRecordAdminJobFailureNotification(ctx, SystemRecordAdminJobFailureNotificationParams{
				RecipientID: 10, ActivityKey: key, LatestAt: occurredAt,
			}); err != nil {
				return err
			}
			return tx.SystemSummarizeAdminJobFailureNotification(ctx, SystemSummarizeAdminJobFailureNotificationParams{
				RecipientID: 10, ActivityKey: key, Summary: "Terminal job failure",
				ReopenBefore: sql.NullTime{Time: occurredAt.Add(-15 * time.Minute), Valid: true},
			})
		}))
	}
	record(now)
	record(now.Add(time.Minute))
	rows, err := q.ListUnreadNotificationsForLister(ctx, ListUnreadNotificationsForListerParams{ListerID: 10, Limit: 10})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Contains(t, rows[0].Message.String, "2 failures")
	require.True(t, rows[0].CreatedAt.Equal(now.Add(time.Minute)))
}

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
