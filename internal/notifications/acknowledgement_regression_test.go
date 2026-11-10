//go:build sqlite || sqlite3

package notifications

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
	_ "modernc.org/sqlite"
)

// setupAckRegressionDB initializes an in-memory SQLite database matching the production schema.
func setupAckRegressionDB(t *testing.T) (*sql.DB, db.Querier) {
	t.Helper()
	conn, err := sql.Open("sqlite", "file:ack_regression?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	schema := `
		CREATE TABLE users (
			idusers INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			email TEXT NOT NULL
		);
		CREATE TABLE notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			users_idusers INTEGER NOT NULL,
			link TEXT,
			message TEXT,
			activity_key TEXT,
			priority INTEGER NOT NULL DEFAULT 0,
			resource_type TEXT,
			resource_id INTEGER,
			activity_position INTEGER,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			read_at DATETIME,
			UNIQUE (users_idusers, activity_key)
		);
		CREATE TABLE content_read_markers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			item TEXT NOT NULL,
			item_id INTEGER NOT NULL,
			last_comment_id INTEGER NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (user_id, item, item_id)
		);
	`
	if _, err := conn.Exec(schema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	queries := db.NewForDriver(conn, "sqlite3")
	return conn, queries
}

// TestRegression_ConcurrentReplies_StaleDisplayedPosition verifies that when a user replies
// using a stale displayed position while a concurrent reply has arrived, the concurrent reply
// notification remains unread.
func TestRegression_ConcurrentReplies_StaleDisplayedPosition(t *testing.T) {
	ctx := context.Background()
	conn, queries := setupAckRegressionDB(t)

	const (
		// recipientID is the user viewing and replying to the thread.
		recipientID int32 = 7
		// threadID is the target thread.
		threadID int32 = 10
		// displayedCommentID is the comment displayed when recipient opened page.
		displayedCommentID int32 = 100
		// concurrentCommentID is a new reply posted by another user concurrently.
		concurrentCommentID int32 = 101
	)

	cfg := config.NewRuntimeConfig()
	notifier := New(WithQueries(queries), WithConfig(cfg))

	// Concurrent reply 101 arrives for recipientID
	pos101 := common.CalculateActivityPosition(concurrentCommentID, 0)
	evt101 := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       "comment:101",
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: pos101,
	}
	if err := notifier.sendInternalNotificationForEvent(ctx, evt101, recipientID, "/forum/thread/10#c101", "concurrent reply", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send concurrent reply notification: %v", err)
	}

	// Verify notification 101 is unread
	var countUnread int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE users_idusers = ? AND read_at IS NULL`, recipientID).Scan(&countUnread); err != nil {
		t.Fatalf("query unread notifications: %v", err)
	}
	if countUnread != 1 {
		t.Fatalf("unread notifications before reply = %d, want 1", countUnread)
	}

	// Recipient submits a reply with stale displayed activity (comment 100)
	pos100 := common.CalculateActivityPosition(displayedCommentID, 0)
	if err := queries.SystemAcknowledgeNotificationsForResource(ctx, db.SystemAcknowledgeNotificationsForResourceParams{
		RecipientID:      recipientID,
		ResourceType:     sql.NullString{String: "thread", Valid: true},
		ResourceID:       sql.NullInt32{Int32: threadID, Valid: true},
		ActivityPosition: sql.NullInt64{Int64: pos100, Valid: true},
	}); err != nil {
		t.Fatalf("acknowledge displayed activity: %v", err)
	}

	// Concurrent notification 101 MUST remain unread because pos100 < pos101
	var unreadAfter int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE users_idusers = ? AND read_at IS NULL`, recipientID).Scan(&unreadAfter); err != nil {
		t.Fatalf("query unread after stale reply: %v", err)
	}
	if unreadAfter != 1 {
		t.Fatalf("unread notifications after reply with stale position = %d, want 1 (must not acknowledge newer concurrent comment)", unreadAfter)
	}
}

// TestRegression_ConsecutiveAppends_Isolation verifies that multiple appends sharing a comment ID
// are acknowledged strictly according to the displayed activity position.
func TestRegression_ConsecutiveAppends_Isolation(t *testing.T) {
	ctx := context.Background()
	conn, queries := setupAckRegressionDB(t)

	const (
		// targetUser is the recipient.
		targetUser int32 = 7
		// threadID is the target thread.
		threadID int32 = 20
		// baseCommentID is the base comment.
		baseCommentID int32 = 200
	)

	notifier := New(WithQueries(queries), WithConfig(config.NewRuntimeConfig()))

	// Base comment 200 arrives
	posBase := common.CalculateActivityPosition(baseCommentID, 0)
	evtBase := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       "comment:200",
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: posBase,
	}
	if err := notifier.sendInternalNotificationForEvent(ctx, evtBase, targetUser, "/forum/thread/20#c200", "base", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send base notification: %v", err)
	}

	// Append 1 arrives
	posApp1 := common.CalculateActivityPosition(baseCommentID, 1)
	evtApp1 := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       "comment:200:append:1",
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: posApp1,
	}
	if err := notifier.sendInternalNotificationForEvent(ctx, evtApp1, targetUser, "/forum/thread/20#c200", "append 1", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send append 1 notification: %v", err)
	}

	// Append 2 arrives
	posApp2 := common.CalculateActivityPosition(baseCommentID, 2)
	evtApp2 := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       "comment:200:append:2",
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: posApp2,
	}
	if err := notifier.sendInternalNotificationForEvent(ctx, evtApp2, targetUser, "/forum/thread/20#c200", "append 2", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send append 2 notification: %v", err)
	}

	checkState := func() map[string]bool {
		rows, err := conn.QueryContext(ctx, `SELECT activity_key, read_at IS NOT NULL FROM notifications WHERE users_idusers = ?`, targetUser)
		if err != nil {
			t.Fatalf("query state: %v", err)
		}
		defer rows.Close()
		m := map[string]bool{}
		for rows.Next() {
			var k string
			var r bool
			if err := rows.Scan(&k, &r); err != nil {
				t.Fatalf("scan state: %v", err)
			}
			m[k] = r
		}
		return m
	}

	// Step 1: User viewed up to base comment
	if err := queries.SystemAcknowledgeNotificationsForResource(ctx, db.SystemAcknowledgeNotificationsForResourceParams{
		RecipientID:      targetUser,
		ResourceType:     sql.NullString{String: "thread", Valid: true},
		ResourceID:       sql.NullInt32{Int32: threadID, Valid: true},
		ActivityPosition: sql.NullInt64{Int64: posBase, Valid: true},
	}); err != nil {
		t.Fatalf("ack base: %v", err)
	}
	s1 := checkState()
	if !s1["reply:comment:200"] || s1["reply:comment:200:append:1"] || s1["reply:comment:200:append:2"] {
		t.Fatalf("state after base ack = %#v, want base read and appends unread", s1)
	}

	// Step 2: User viewed up to append 1
	if err := queries.SystemAcknowledgeNotificationsForResource(ctx, db.SystemAcknowledgeNotificationsForResourceParams{
		RecipientID:      targetUser,
		ResourceType:     sql.NullString{String: "thread", Valid: true},
		ResourceID:       sql.NullInt32{Int32: threadID, Valid: true},
		ActivityPosition: sql.NullInt64{Int64: posApp1, Valid: true},
	}); err != nil {
		t.Fatalf("ack append 1: %v", err)
	}
	s2 := checkState()
	if !s2["reply:comment:200"] || !s2["reply:comment:200:append:1"] || s2["reply:comment:200:append:2"] {
		t.Fatalf("state after append 1 ack = %#v, want base and append 1 read, append 2 unread", s2)
	}

	// Step 3: User viewed up to append 2
	if err := queries.SystemAcknowledgeNotificationsForResource(ctx, db.SystemAcknowledgeNotificationsForResourceParams{
		RecipientID:      targetUser,
		ResourceType:     sql.NullString{String: "thread", Valid: true},
		ResourceID:       sql.NullInt32{Int32: threadID, Valid: true},
		ActivityPosition: sql.NullInt64{Int64: posApp2, Valid: true},
	}); err != nil {
		t.Fatalf("ack append 2: %v", err)
	}
	s3 := checkState()
	if !s3["reply:comment:200"] || !s3["reply:comment:200:append:1"] || !s3["reply:comment:200:append:2"] {
		t.Fatalf("state after append 2 ack = %#v, want all read", s3)
	}
}

// TestRegression_ReadAfterDelivery verifies that normal notification delivery followed by user read
// transitions the notification from unread to read.
func TestRegression_ReadAfterDelivery(t *testing.T) {
	ctx := context.Background()
	conn, queries := setupAckRegressionDB(t)

	const (
		// userA is the recipient.
		userA int32 = 7
		// threadID is the target thread.
		threadID int32 = 30
		// commentID is the delivered reply.
		commentID int32 = 300
	)

	notifier := New(WithQueries(queries), WithConfig(config.NewRuntimeConfig()))

	// Notification delivered to userA
	pos := common.CalculateActivityPosition(commentID, 0)
	evt := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       "comment:300",
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: pos,
	}
	if err := notifier.sendInternalNotificationForEvent(ctx, evt, userA, "/forum/thread/30#c300", "delivered reply", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send notification: %v", err)
	}

	// Verify unread
	var readAt sql.NullTime
	if err := conn.QueryRowContext(ctx, `SELECT read_at FROM notifications WHERE users_idusers = ? AND activity_key = ?`, userA, "reply:comment:300").Scan(&readAt); err != nil {
		t.Fatalf("scan notification: %v", err)
	}
	if readAt.Valid {
		t.Fatalf("notification should be unread initially, got read_at = %v", readAt.Time)
	}

	// User reads thread up to comment 300
	if err := queries.SystemAcknowledgeNotificationsForResource(ctx, db.SystemAcknowledgeNotificationsForResourceParams{
		RecipientID:      userA,
		ResourceType:     sql.NullString{String: "thread", Valid: true},
		ResourceID:       sql.NullInt32{Int32: threadID, Valid: true},
		ActivityPosition: sql.NullInt64{Int64: pos, Valid: true},
	}); err != nil {
		t.Fatalf("acknowledge notification: %v", err)
	}

	// Verify read
	if err := conn.QueryRowContext(ctx, `SELECT read_at FROM notifications WHERE users_idusers = ? AND activity_key = ?`, userA, "reply:comment:300").Scan(&readAt); err != nil {
		t.Fatalf("scan notification after ack: %v", err)
	}
	if !readAt.Valid {
		t.Fatalf("notification should be marked read after acknowledgement")
	}
}

// TestRegression_DeliveryAfterRead verifies that late notification arrival for an already-read
// activity position is immediately marked read, while late arrivals for unseen appends or newer
// comments remain unread.
func TestRegression_DeliveryAfterRead(t *testing.T) {
	ctx := context.Background()
	conn, queries := setupAckRegressionDB(t)

	const (
		// userA is the recipient.
		userA int32 = 7
		// threadID is the target thread.
		threadID int32 = 40
		// readCommentID is the comment up to which the user has already read.
		readCommentID int32 = 400
	)

	// User has read marker up to comment 400
	if err := queries.UpsertContentReadMarker(ctx, db.UpsertContentReadMarkerParams{
		Item:          "thread",
		ItemID:        threadID,
		UserID:        userA,
		LastCommentID: readCommentID,
	}); err != nil {
		t.Fatalf("upsert read marker: %v", err)
	}

	notifier := New(WithQueries(queries), WithConfig(config.NewRuntimeConfig()))

	// Late delivery 1: Notification for comment 400 (base) arrives late.
	// Since user read marker is 400, this should be marked read immediately upon creation.
	pos400 := common.CalculateActivityPosition(readCommentID, 0)
	evt400 := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       "comment:400",
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: pos400,
	}
	if err := notifier.sendInternalNotificationForEvent(ctx, evt400, userA, "/forum/thread/40#c400", "late comment 400", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send late comment 400: %v", err)
	}

	// Late delivery 2: Notification for comment 399 (older comment) arrives late.
	// Since user read marker is 400 > 399, this should also be marked read immediately.
	pos399 := common.CalculateActivityPosition(399, 0)
	evt399 := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       "comment:399",
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: pos399,
	}
	if err := notifier.sendInternalNotificationForEvent(ctx, evt399, userA, "/forum/thread/40#c399", "late comment 399", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send late comment 399: %v", err)
	}

	// Late delivery 3: Notification for comment 401 (newer comment) arrives.
	// Since user read marker is 400 < 401, this MUST be unread.
	pos401 := common.CalculateActivityPosition(401, 0)
	evt401 := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       "comment:401",
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: pos401,
	}
	if err := notifier.sendInternalNotificationForEvent(ctx, evt401, userA, "/forum/thread/40#c401", "new comment 401", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send new comment 401: %v", err)
	}

	// Late delivery 4: Notification for comment 400:append:1 arrives.
	// Since user read marker was at comment 400, the user has not seen append 1.
	// It MUST be unread.
	pos400App1 := common.CalculateActivityPosition(readCommentID, 1)
	evt400App1 := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       "comment:400:append:1",
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: pos400App1,
	}
	if err := notifier.sendInternalNotificationForEvent(ctx, evt400App1, userA, "/forum/thread/40#c400", "append to 400", notificationPrioritySubscriber); err != nil {
		t.Fatalf("send append to 400: %v", err)
	}

	// Verify states
	var read400, read399, read401, read400App1 sql.NullTime
	_ = conn.QueryRowContext(ctx, `SELECT read_at FROM notifications WHERE users_idusers = ? AND activity_key = ?`, userA, "reply:comment:400").Scan(&read400)
	_ = conn.QueryRowContext(ctx, `SELECT read_at FROM notifications WHERE users_idusers = ? AND activity_key = ?`, userA, "reply:comment:399").Scan(&read399)
	_ = conn.QueryRowContext(ctx, `SELECT read_at FROM notifications WHERE users_idusers = ? AND activity_key = ?`, userA, "reply:comment:401").Scan(&read401)
	_ = conn.QueryRowContext(ctx, `SELECT read_at FROM notifications WHERE users_idusers = ? AND activity_key = ?`, userA, "reply:comment:400:append:1").Scan(&read400App1)

	if !read400.Valid {
		t.Errorf("comment 400 should be marked read immediately upon late delivery")
	}
	if !read399.Valid {
		t.Errorf("comment 399 should be marked read immediately upon late delivery")
	}
	if read401.Valid {
		t.Errorf("comment 401 should remain unread, got read_at = %v", read401.Time)
	}
	if read400App1.Valid {
		t.Errorf("comment 400:append:1 should remain unread because user only read base comment, got read_at = %v", read400App1.Time)
	}
}
