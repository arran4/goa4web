//go:build !sqlite && !sqlite3

package db

import (
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/arran4/goa4web/database"
)

func TestMySQLForumReplyEmailWindowClaimAndEligibility(t *testing.T) {
	ctx := t.Context()
	conn := openReplyThreadTestDatabase(t)
	if _, err := conn.ExecContext(ctx, string(database.SchemaMySQL)); err != nil {
		t.Fatalf("load canonical schema: %v", err)
	}
	execReplySQL(t, conn,
		"INSERT INTO users (idusers, username) VALUES (1, 'viewer'), (2, 'author')",
		"INSERT INTO preferences (users_idusers, emailforumupdates) VALUES (1, 1)",
		"INSERT INTO forumtopic (idforumtopic, handler, title) VALUES (5, 'private', 'Example topic')",
		"INSERT INTO forumthread (idforumthread, forumtopic_idforumtopic, firstpost) VALUES (6, 5, 7)",
		"INSERT INTO comments (idcomments, forumthread_id, users_idusers, text) VALUES (7, 6, 2, 'opening'), (8, 6, 2, 'reply')",
		"INSERT INTO grants (user_id, section, item, item_id, action, rule_type) VALUES (1, 'privateforum', 'topic', 5, 'view', 'allow'), (1, 'privateforum_thread', 'thread', 6, 'view', 'allow')",
	)
	q := NewForDriver(conn, "mysql")
	activity := SystemQueueForumReplyEmailActivityParams{RecipientID: 1, ThreadID: 6, ActivityPosition: 8 << 32, WindowEnd: 100}
	for range 2 {
		if err := q.SystemQueueForumReplyEmailActivity(ctx, activity); err != nil {
			t.Fatal(err)
		}
	}
	activity.ActivityPosition++
	if err := q.SystemQueueForumReplyEmailActivity(ctx, activity); err != nil {
		t.Fatal(err)
	}

	due := SystemGetDueForumReplyEmailWindowParams{DueAt: 99, LeaseAvailableAt: 99}
	if _, err := q.SystemGetDueForumReplyEmailWindow(ctx, due); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("window found before it closed: %v", err)
	}
	due.DueAt, due.LeaseAvailableAt = 100, 100
	window, err := q.SystemGetDueForumReplyEmailWindow(ctx, due)
	if err != nil || window.RecipientID != 1 || window.WindowEnd != 100 {
		t.Fatalf("expected due window, got %+v, %v", window, err)
	}

	// Competing worker updates may observe the same due window. Exactly one
	// may claim its activity rows, and the loser must never dispatch them.
	type claimResult struct {
		token sql.NullString
		count int64
		err   error
	}
	results := make(chan claimResult, 2)
	var start sync.WaitGroup
	start.Add(1)
	for _, name := range []string{"worker-one", "worker-two"} {
		go func() {
			token := sql.NullString{String: name, Valid: true}
			start.Wait()
			count, err := q.SystemClaimForumReplyEmailWindow(ctx, SystemClaimForumReplyEmailWindowParams{
				RecipientID: 1, WindowEnd: 100, LeaseToken: token, LeaseAvailableAt: 100, LeaseUntil: 220,
			})
			results <- claimResult{token: token, count: count, err: err}
		}()
	}
	start.Done()
	var winner sql.NullString
	for range 2 {
		res := <-results
		if res.err != nil {
			t.Fatal(res.err)
		}
		if res.count == 2 && !winner.Valid {
			winner = res.token
		} else if res.count != 0 {
			t.Fatalf("split or duplicate claim: %+v", res)
		}
	}
	if !winner.Valid {
		t.Fatal("neither worker claimed the window")
	}
	args := SystemListEligibleForumReplyEmailThreadsParams{RecipientID: 1, WindowEnd: 100, LeaseToken: winner, Limit: 25}
	rows, err := q.SystemListEligibleForumReplyEmailThreads(ctx, args)
	if err != nil || len(rows) != 1 || rows[0].ReplyCount != 2 {
		t.Fatalf("grouped original and append, got %v, error %v", rows, err)
	}
	execReplySQL(t, conn, "DELETE FROM grants WHERE section = 'privateforum_thread'")
	rows, err = q.SystemListEligibleForumReplyEmailThreads(ctx, args)
	if err != nil || len(rows) != 0 {
		t.Fatalf("revoked thread returned %d rows, error %v", len(rows), err)
	}
}
