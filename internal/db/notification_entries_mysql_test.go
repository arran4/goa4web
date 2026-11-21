//go:build !sqlite && !sqlite3

package db

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/arran4/goa4web/database"
)

func TestMySQLNotificationEntriesRespectGrantsAndGroupSnapshot(t *testing.T) {
	ctx := t.Context()
	conn := openReplyThreadTestDatabase(t)
	if _, err := conn.ExecContext(ctx, string(database.SchemaMySQL)); err != nil {
		t.Fatalf("load canonical schema: %v", err)
	}

	// This is a query contract fixture. The application scenario separately
	// creates its replies and notifications through the normal command path.
	execReplySQL(t, conn,
		"INSERT INTO users (idusers, username) VALUES (1, 'viewer'), (2, 'author')",
		"INSERT INTO forumtopic (idforumtopic, handler, title) VALUES (5, 'private', 'Example topic')",
		"INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic, firstpost) VALUES (6, 5, 7)",
		"INSERT INTO comments (idcomments, forumthread_id, users_idusers, text) VALUES (7, 6, 2, 'opening'), (8, 6, 2, 'reply')",
		"INSERT INTO grants (user_id, section, item, item_id, action, rule_type) VALUES (1, 'privateforum', 'topic', 5, 'view', 'allow'), (1, 'privateforum_thread', 'thread', 6, 'view', 'allow')",
	)
	q := NewForDriver(conn, "mysql")
	for position, key := range map[int64]string{int64(8) << 32: "reply:comment:8", int64(8)<<32 | 1: "reply:comment:8:append:1"} {
		arg := SystemCreateNotificationForActivityParams{
			RecipientID: 1, ActivityKey: sql.NullString{String: key, Valid: true},
			ResourceType: sql.NullString{String: "thread", Valid: true}, ResourceID: sql.NullInt32{Int32: 6, Valid: true},
			ActivityPosition: sql.NullInt64{Int64: position, Valid: true},
		}
		if err := q.SystemCreateNotificationForActivity(ctx, arg); err != nil {
			t.Fatal(err)
		}
	}
	arg := ListVisibleNotificationEntriesForListerParams{ListerID: 1, GroupUnread: 1, IncludeRead: false, Limit: 10}
	rows, err := q.ListVisibleNotificationEntriesForLister(ctx, arg)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ReplyCount != 2 || rows[0].TotalEntries != 1 || rows[0].ReplyAuthor != "author" {
		t.Fatalf("expected one entry with both replies, got %+v", rows)
	}
	var raw []byte
	switch snapshot := rows[0].Snapshot.(type) {
	case []byte:
		raw = snapshot
	case string:
		raw = []byte(snapshot)
	}
	var ids []int32
	if err := json.Unmarshal(raw, &ids); err != nil || len(ids) != 2 {
		t.Fatalf("snapshot %s, error %v, expected two member IDs", raw, err)
	}
	if _, err := conn.ExecContext(ctx, "DELETE FROM grants WHERE section = 'privateforum_thread' AND item_id = 6"); err != nil {
		t.Fatal(err)
	}
	rows, err = q.ListVisibleNotificationEntriesForLister(ctx, arg)
	if err != nil || len(rows) != 0 {
		t.Fatalf("revoked entry count = %d, err = %v", len(rows), err)
	}
}

func TestMySQLNotificationObservationIsScopedToActiveLister(t *testing.T) {
	ctx := t.Context()
	conn := openReplyThreadTestDatabase(t)
	if _, err := conn.ExecContext(ctx, string(database.SchemaMySQL)); err != nil {
		t.Fatalf("load canonical schema: %v", err)
	}
	execReplySQL(t, conn,
		"INSERT INTO users (idusers, username) VALUES (1, 'viewer'), (2, 'other viewer')",
		"INSERT INTO roles (id, name, can_login) VALUES (1, 'member', 1)",
		"INSERT INTO user_roles (users_idusers, role_id) VALUES (1, 1), (2, 1)",
		"INSERT INTO notifications (id, users_idusers, message) VALUES (1, 1, 'first'), (2, 1, 'second'), (3, 2, 'other')",
	)
	q := NewForDriver(conn, "mysql")
	id, err := q.GetLatestNotificationIDForLister(ctx, 1)
	if err != nil || id != 2 {
		t.Fatalf("viewer latest ID = %d, error = %v, want 2", id, err)
	}
	rows, err := q.ListVisibleNotificationEntriesForLister(ctx, ListVisibleNotificationEntriesForListerParams{
		ListerID: 1, AfterNotificationID: 1, IncludeRead: true, Limit: 10,
	})
	if err != nil || len(rows) != 1 || rows[0].ID != 2 {
		t.Fatalf("viewer newer notifications = %+v, error = %v", rows, err)
	}
	execReplySQL(t, conn, "DELETE FROM user_roles WHERE users_idusers = 1")
	if _, err := q.GetLatestNotificationIDForLister(ctx, 1); err != sql.ErrNoRows {
		t.Fatalf("viewer without login role: error = %v, want sql.ErrNoRows", err)
	}
}
