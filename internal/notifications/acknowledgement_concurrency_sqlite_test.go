//go:build sqlite || sqlite3

package notifications

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
	_ "modernc.org/sqlite"
)

func TestNotifierAndAcknowledge_DeterministicBarrier_SQLite(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open SQLite database: %v", err)
	}
	conn.SetMaxOpenConns(4)
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.ExecContext(ctx, `
		CREATE TABLE content_read_markers (
			id integer PRIMARY KEY AUTOINCREMENT,
			item TEXT NOT NULL,
			item_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			last_comment_id INTEGER NOT NULL,
			last_activity_position INTEGER NOT NULL DEFAULT 0,
			UNIQUE (user_id, item_id, item)
		);
		CREATE TABLE notifications (
			id integer PRIMARY KEY AUTOINCREMENT,
			users_idusers integer NOT NULL,
			link TEXT,
			message TEXT,
			activity_key TEXT,
			priority integer NOT NULL DEFAULT 0,
			resource_type TEXT,
			resource_id integer,
			activity_position integer,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			read_at TIMESTAMP DEFAULT NULL,
			UNIQUE (users_idusers, activity_key)
		);
	`); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	queries := db.NewForDriver(conn, "sqlite3")
	n := &Notifier{Queries: queries, Config: config.NewRuntimeConfig(), Silent: true}

	const (
		recipientID int32 = 12
		threadID    int32 = 88
		commentID   int32 = 400
		appendID    int64 = 1
	)
	pos := common.CalculateActivityPosition(commentID, appendID)
	evt := eventbus.TaskEvent{
		Task:             tasks.TaskString("Reply"),
		ActivityID:       common.FormatActivityPosition(commentID, appendID),
		ResourceType:     "thread",
		ResourceID:       threadID,
		ActivityPosition: pos,
	}

	deliverAtBarrier := make(chan struct{})
	ackStarted := make(chan struct{})
	ackErrCh := make(chan error, 1)
	deliverErrCh := make(chan error, 1)

	deliverCtx := db.WithReadStateTxHooks(ctx, db.ReadStateTxHooks{
		AfterDeliverMarkerRead: func(context.Context) {
			close(deliverAtBarrier)
			select {
			case <-ackStarted:
			case <-ctx.Done():
			}
			time.Sleep(15 * time.Millisecond)
		},
	})

	go func() {
		deliverErrCh <- n.sendInternalNotificationForEvent(deliverCtx, evt, recipientID, "/private/topic/1/thread/88#c400", "append 1", notificationPrioritySubscriber)
	}()

	<-deliverAtBarrier
	go func() {
		cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserID(recipientID))
		close(ackStarted)
		ackErrCh <- cd.AcknowledgeThreadActivityPosition(threadID, commentID, pos)
	}()

	if err := <-deliverErrCh; err != nil {
		t.Fatalf("sendInternalNotificationForEvent: %v", err)
	}
	if err := <-ackErrCh; err != nil {
		t.Fatalf("AcknowledgeThreadActivityPosition: %v", err)
	}

	unread, err := queries.GetUnreadNotificationCountForLister(ctx, recipientID)
	if err != nil {
		t.Fatalf("GetUnreadNotificationCountForLister: %v", err)
	}
	if unread != 0 {
		t.Fatalf("expected 0 unread notifications after racing Notifier and AcknowledgeThreadActivityPosition, got %d", unread)
	}
}
