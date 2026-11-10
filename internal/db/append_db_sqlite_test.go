//go:build sqlite || sqlite3

package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestAppendCommentSQLite(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1'), (2, 'user2')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', datetime('now', '-30 minutes'))`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 100, 'append', 1, 'allow')`,
	)

	queries := NewForDriver(database, "sqlite")
	ctx := context.Background()
	now := time.Now()

	appendParams := AppendCommentInSectionForCommenterParams{
		Text:             "new text sqlite",
		AppendWindowMins: 60,
		Section:          "forum",
		ItemType:         sql.NullString{String: "topic", Valid: true},
		GrantUserID:      sql.NullInt32{Int32: 1, Valid: true},
		ItemID:           sql.NullInt32{Int32: 100, Valid: true},
		Written:          sql.NullTime{Time: now, Valid: true},
		CommentID:        1000,
		CommenterID:      1,
		ForumthreadID:    100,
	}
	rowsAffected, err := queries.AppendCommentInSectionForCommenter(ctx, appendParams)

	if err != nil {
		t.Fatalf("append comment: %v", err)
	}

	if rowsAffected != 1 {
		t.Fatalf("expected 1 row affected, got %d", rowsAffected)
	}

	comment, err := queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}

	expectedText := "initial text\n\n[hr]\n\nnew text sqlite"
	if comment.Text.String != expectedText {
		t.Errorf("expected text %q, got %q", expectedText, comment.Text.String)
	}
	if !comment.Written.Valid || comment.Written.Time.Sub(now).Abs() > time.Second {
		t.Errorf("written = %v, want %v", comment.Written, now)
	}
	var commentCount int
	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", 100).Scan(&commentCount); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if commentCount != 1 || comment.Idcomments != 1000 {
		t.Fatalf("append changed comment identity/count: id=%d count=%d", comment.Idcomments, commentCount)
	}

	appendParams.Text = "second segment"
	appendParams.Written = sql.NullTime{Time: now.Add(time.Second), Valid: true}
	rowsAffected, err = queries.AppendCommentInSectionForCommenter(ctx, appendParams)
	if err != nil {
		t.Fatalf("repeat append comment: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("expected repeat append to affect 1 row, got %d", rowsAffected)
	}
	comment, err = queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get repeatedly appended comment: %v", err)
	}
	expectedText += "\n\n[hr]\n\nsecond segment"
	if comment.Text.String != expectedText {
		t.Fatalf("repeated append text = %q, want %q", comment.Text.String, expectedText)
	}
	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE forumthread_id = ?", 100).Scan(&commentCount); err != nil {
		t.Fatalf("count comments after repeat append: %v", err)
	}
	if commentCount != 1 || comment.Idcomments != 1000 {
		t.Fatalf("repeat append changed comment identity/count: id=%d count=%d", comment.Idcomments, commentCount)
	}
}

func TestAppendCommentWithNullExistingTextSQLite(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)
	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, NULL, datetime('now'))`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 100, 'append', 1, 'allow')`,
	)
	queries := NewForDriver(database, "sqlite")
	ctx := context.Background()

	rowsAffected, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
		Text: "new text", AppendWindowMins: 60, Section: "forum",
		ItemType: sql.NullString{String: "topic", Valid: true}, GrantUserID: sql.NullInt32{Int32: 1, Valid: true},
		ItemID: sql.NullInt32{Int32: 100, Valid: true}, Written: sql.NullTime{Time: time.Now(), Valid: true},
		CommentID: 1000, CommenterID: 1, ForumthreadID: 100,
	})
	if err != nil {
		t.Fatalf("append comment with NULL text: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("expected 1 row affected, got %d", rowsAffected)
	}
	comment, err := queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}
	if got, want := comment.Text.String, "\n\n[hr]\n\nnew text"; got != want || !comment.Text.Valid {
		t.Fatalf("NULL append text = %#v, want %q", comment.Text, want)
	}
}

func TestAppendCommentConcurrencySQLite(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1'), (2, 'user2')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', datetime('now', '-30 minutes'))`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 100, 'append', 1, 'allow')`,
	)

	queries := NewForDriver(database, "sqlite")
	ctx := context.Background()

	// Simulate concurrent append attempts
	ch1 := make(chan error)
	ch2 := make(chan error)

	now := time.Now()

	go func() {
		_, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
			Text:             "append 1",
			AppendWindowMins: 60,
			Section:          "forum",
			ItemType:         sql.NullString{String: "topic", Valid: true},
			GrantUserID:      sql.NullInt32{Int32: 1, Valid: true},
			ItemID:           sql.NullInt32{Int32: 100, Valid: true},
			Written:          sql.NullTime{Time: now, Valid: true},
			CommentID:        1000,
			CommenterID:      1,
			ForumthreadID:    100,
		})
		ch1 <- err
	}()

	go func() {
		_, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
			Text:             "append 2",
			AppendWindowMins: 60,
			Section:          "forum",
			ItemType:         sql.NullString{String: "topic", Valid: true},
			GrantUserID:      sql.NullInt32{Int32: 1, Valid: true},
			ItemID:           sql.NullInt32{Int32: 100, Valid: true},
			Written:          sql.NullTime{Time: now, Valid: true},
			CommentID:        1000,
			CommenterID:      1,
			ForumthreadID:    100,
		})
		ch2 <- err
	}()

	err1 := <-ch1
	err2 := <-ch2

	if err1 != nil {
		t.Fatalf("append 1: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("append 2: %v", err2)
	}

	comment, err := queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}

	finalText := comment.Text.String

	if !strings.Contains(finalText, "append 1") {
		t.Errorf("missing append 1 in: %q", finalText)
	}
	if !strings.Contains(finalText, "append 2") {
		t.Errorf("missing append 2 in: %q", finalText)
	}
}

func TestCommentAppendAtomicIdentityConcurrencySQLite(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', datetime('now', '-30 minutes'))`,
	)

	queries := NewForDriver(database, "sqlite")
	ctx := context.Background()

	// Simulate concurrent append identity allocation
	// numConcurrent specifies the number of parallel append operations.
	const numConcurrent = 5
	appendIDs := make(chan int64, numConcurrent)
	errCh := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func() {
			id, err := queries.SystemCreateCommentAppend(ctx, SystemCreateCommentAppendParams{
				CommentID:    1000,
				UsersIdusers: 1,
				CreatedAt:    time.Now().UTC(),
			})
			if err != nil {
				errCh <- err
				return
			}
			appendIDs <- id
		}()
	}

	seen := make(map[int64]bool)
	for i := 0; i < numConcurrent; i++ {
		select {
		case err := <-errCh:
			t.Fatalf("SystemCreateCommentAppend concurrent error: %v", err)
		case id := <-appendIDs:
			if id <= 0 {
				t.Fatalf("expected positive auto-increment append ID, got %d", id)
			}
			if seen[id] {
				t.Fatalf("duplicate append ID %d allocated under concurrency", id)
			}
			seen[id] = true
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent append IDs")
		}
	}

	appends, err := queries.SystemListCommentAppendsByCommentID(ctx, 1000)
	if err != nil {
		t.Fatalf("SystemListCommentAppendsByCommentID: %v", err)
	}
	if len(appends) != numConcurrent {
		t.Fatalf("recorded appends count = %d, want %d", len(appends), numConcurrent)
	}
	for i := 1; i < len(appends); i++ {
		if appends[i].ID <= appends[i-1].ID {
			t.Fatalf("expected monotonically increasing IDs, got %d then %d", appends[i-1].ID, appends[i].ID)
		}
	}
}

func openReplyThreadTestDatabaseSQLite(t *testing.T) *sql.DB {
	t.Helper()
	// Using file::memory:?cache=shared instead of just :memory: for concurrent accesses
	db, err := sql.Open("sqlite", "file:"+t.Name()+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func mustExec(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("setup exec failed: %v", err)
	}
}

func testAppendEligibilityMatrix(t *testing.T, database *sql.DB, queries Querier, ctx context.Context) {
	// Instead of using real queries (which might be missing in Querier stub or generated types),
	// we will manually setup the DB via exec.

	// Create user
	mustExec(t, ctx, database, "INSERT INTO users (idusers, username) VALUES (?, ?)", 60, "poster")
	mustExec(t, ctx, database, "INSERT INTO users (idusers, username) VALUES (?, ?)", 61, "reader")

	// Create thread
	mustExec(t, ctx, database, "INSERT INTO forumthread (idforumthread, firstpost, lastposter, forumtopic_idforumtopic) VALUES (?, ?, ?, ?)", 1000, 2000, 60, 30)

	// Create comment
	mustExec(t, ctx, database, "INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (?, ?, ?, ?, '2030-01-01 12:00:00')", 2000, 1000, 60, "Initial post")
	mustExec(t, ctx, database, "UPDATE forumthread SET firstpost = 2000, lastposter = 60 WHERE idforumthread = 1000")

	// Create grant for poster
	mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "forum", "topic", 30, "append", 1, "allow")

	tests := []struct {
		name     string
		setup    func()
		section  string
		itemType string
		itemID   int32
		wantRows bool
	}{
		{
			name:    "no marker -> 1 row",
			section: "forum", itemType: "topic", itemID: 30,
			setup:    func() {},
			wantRows: true,
		},
		{
			name:    "author's own marker -> 1 row",
			section: "forum", itemType: "topic", itemID: 30,
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 60, 2000)
			},
			wantRows: true,
		},
		{
			name:    "other user's older marker -> 1 row",
			section: "forum", itemType: "topic", itemID: 30,
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 61, 1999)
			},
			wantRows: true,
		},
		{
			name:    "other user's marker at comment -> 0 rows",
			section: "forum", itemType: "topic", itemID: 30,
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 61, 2000)
			},
			wantRows: false,
		},
		{
			name: "other user's marker beyond -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO content_read_markers (item, item_id, user_id, last_comment_id) VALUES (?, ?, ?, ?)", "thread", 1000, 61, 2001)
			},
			wantRows: false,
		},
		{
			name: "newer comment exists -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (?, ?, ?, ?, '2030-01-01 12:01:00')", 2001, 1000, 60, "newer")
			},
			wantRows: false,
		},
		{
			name: "wrong owner -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "UPDATE comments SET users_idusers = 61 WHERE idcomments = 2000")
			},
			wantRows: false,
		},
		{
			name: "append grant missing -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "DELETE FROM grants")
			},
			wantRows: false,
		},
		{
			name: "outside window -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "UPDATE comments SET written = '2000-01-01 12:00:00' WHERE idcomments = 2000")
			},
			wantRows: false,
		},
		{
			name: "privateforum_thread/thread grant -> 1 row",
			setup: func() {

				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "privateforum_thread", "thread", 1000, "append", 1, "allow")
			},
			section:  "privateforum_thread",
			itemType: "thread",
			itemID:   1000,
			wantRows: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mustExec(t, ctx, database, "DELETE FROM comments WHERE idcomments > 2000")
			mustExec(t, ctx, database, "DELETE FROM content_read_markers")
			mustExec(t, ctx, database, "DELETE FROM grants")
			mustExec(t, ctx, database, "UPDATE comments SET users_idusers = 60, text = 'Initial post', written = '2030-01-01 12:00:00' WHERE idcomments = 2000")

			section := "forum"
			if tc.section == "privateforum_thread" {
				section = tc.section
			}
			if section == "forum" {
				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "forum", "topic", 30, "append", 1, "allow")
			}

			tc.setup()

			rowsAffected, err := queries.AppendCommentInSectionForCommenter(ctx, AppendCommentInSectionForCommenterParams{
				Text:             "new text",
				AppendWindowMins: 60, // normal window
				Section: func() string {
					if tc.section == "" {
						return "forum"
					} else {
						return tc.section
					}
				}(),
				ItemType: sql.NullString{String: func() string {
					if tc.itemType == "" {
						return "topic"
					} else {
						return tc.itemType
					}
				}(), Valid: true},
				ItemID: sql.NullInt32{Int32: func() int32 {
					if tc.itemID == 0 {
						return 30
					} else {
						return tc.itemID
					}
				}(), Valid: true},
				CommentID:     2000,
				CommenterID:   60,
				ForumthreadID: 1000,
				GrantUserID:   sql.NullInt32{Int32: 60, Valid: true},
				Written:       sql.NullTime{Time: time.Now(), Valid: true},
			})
			hasAppended := rowsAffected > 0
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			if hasAppended != tc.wantRows {
				t.Fatalf("Expected wantRows=%v, got=%v", tc.wantRows, hasAppended)
			}
			if tc.wantRows {
				cRow, _ := queries.GetCommentById(ctx, 2000)
				if cRow.Text.String != "Initial post\n\n[hr]\n\nnew text" {
					t.Fatalf("Expected exact text, got: %q", cRow.Text.String)
				}
			}
		})
	}
}

func TestSQLiteAppendEligibilityMatrix(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)
	queries := NewForDriver(database, "sqlite")
	ctx := context.Background()
	testAppendEligibilityMatrix(t, database, queries, ctx)
}

func TestAppendCommentInSectionWithActivity_InsertFailureRollbackSQLite(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', datetime('now', '-10 minutes'))`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 100, 'append', 1, 'allow')`,
	)

	queries := NewForDriver(database, "sqlite")
	appender, ok := queries.(CommentAppender)
	if !ok {
		t.Fatalf("queries does not implement CommentAppender")
	}
	ctx := context.Background()
	now := time.Now().UTC()

	appendParams := AppendCommentInSectionForCommenterParams{
		Text:             "new text sqlite",
		AppendWindowMins: 60,
		Section:          "forum",
		ItemType:         sql.NullString{String: "topic", Valid: true},
		GrantUserID:      sql.NullInt32{Int32: 1, Valid: true},
		ItemID:           sql.NullInt32{Int32: 100, Valid: true},
		Written:          sql.NullTime{Time: now, Valid: true},
		CommentID:        1000,
		CommenterID:      1,
		ForumthreadID:    100,
	}

	// Step 1: Install a trigger that aborts any INSERT into comment_appends to simulate a ledger write failure.
	execReplySQL(t, database,
		`CREATE TRIGGER fail_comment_appends BEFORE INSERT ON comment_appends BEGIN SELECT RAISE(ABORT, 'injected insert failure'); END;`,
	)

	rowsAffected, appendID, err := appender.AppendCommentInSectionWithActivity(ctx, appendParams)
	if err == nil {
		t.Fatal("expected AppendCommentInSectionWithActivity to fail when comment_appends INSERT fails, but got nil error")
	}
	if !strings.Contains(err.Error(), "injected insert failure") && !strings.Contains(err.Error(), "record comment append") {
		t.Fatalf("unexpected error message: %v", err)
	}
	if rowsAffected != 0 || appendID != 0 {
		t.Fatalf("expected 0 rows affected and 0 appendID on failure, got rows=%d appendID=%d", rowsAffected, appendID)
	}

	// Verify atomic rollback: the comments row must still contain the unmutated 'initial text'.
	comment, err := queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get comment after rollback: %v", err)
	}
	if comment.Text.String != "initial text" {
		t.Fatalf("expected comment text to remain 'initial text' after rollback, got %q", comment.Text.String)
	}

	// Verify no orphaned records exist in comment_appends.
	var appendCount int
	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM comment_appends WHERE comment_id = ?", 1000).Scan(&appendCount); err != nil {
		t.Fatalf("count comment_appends: %v", err)
	}
	if appendCount != 0 {
		t.Fatalf("expected 0 comment_appends rows after rollback, got %d", appendCount)
	}

	// Step 2: Remove the failure trigger and retry the exact same logical operation.
	execReplySQL(t, database, `DROP TRIGGER fail_comment_appends;`)

	rowsAffected, appendID, err = appender.AppendCommentInSectionWithActivity(ctx, appendParams)
	if err != nil {
		t.Fatalf("retry append after removing failure trigger: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("expected 1 row affected on retry, got %d", rowsAffected)
	}
	if appendID <= 0 {
		t.Fatalf("expected positive append ID on retry, got %d", appendID)
	}

	// Verify successful commit: comment text is updated and append ledger contains 1 record.
	comment, err = queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get comment after retry: %v", err)
	}
	expectedText := "initial text\n\n[hr]\n\nnew text sqlite"
	if comment.Text.String != expectedText {
		t.Fatalf("expected comment text %q, got %q", expectedText, comment.Text.String)
	}

	appends, err := queries.SystemListCommentAppendsByCommentID(ctx, 1000)
	if err != nil {
		t.Fatalf("list comment appends: %v", err)
	}
	if len(appends) != 1 || int64(appends[0].ID) != appendID {
		t.Fatalf("expected 1 append record with ID %d, got %+v", appendID, appends)
	}
}

func TestAppendCommentInSectionWithActivity_ZeroRowsAffectedNoActivitySQLite(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1')`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (1000, 100, 1, 'initial text', datetime('now', '-30 minutes'))`,
	)

	queries := NewForDriver(database, "sqlite")
	appender, ok := queries.(CommentAppender)
	if !ok {
		t.Fatalf("queries does not implement CommentAppender")
	}
	ctx := context.Background()

	// AppendWindowMins is 10, but comment was written 30 minutes ago: UPDATE matches 0 rows.
	appendParams := AppendCommentInSectionForCommenterParams{
		Text:             "should not append",
		AppendWindowMins: 10,
		Section:          "forum",
		ItemType:         sql.NullString{String: "topic", Valid: true},
		GrantUserID:      sql.NullInt32{Int32: 1, Valid: true},
		ItemID:           sql.NullInt32{Int32: 100, Valid: true},
		Written:          sql.NullTime{Time: time.Now().UTC(), Valid: true},
		CommentID:        1000,
		CommenterID:      1,
		ForumthreadID:    100,
	}

	rowsAffected, appendID, err := appender.AppendCommentInSectionWithActivity(ctx, appendParams)
	if err != nil {
		t.Fatalf("unexpected error on 0 rows affected: %v", err)
	}
	if rowsAffected != 0 || appendID != 0 {
		t.Fatalf("expected rows=0 appendID=0, got rows=%d appendID=%d", rowsAffected, appendID)
	}

	// Verify no comment_appends record was created.
	var appendCount int
	if err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM comment_appends WHERE comment_id = ?", 1000).Scan(&appendCount); err != nil {
		t.Fatalf("count comment_appends: %v", err)
	}
	if appendCount != 0 {
		t.Fatalf("expected 0 comment_appends rows when UPDATE affects 0 rows, got %d", appendCount)
	}

	// Verify comment text was unchanged.
	comment, err := queries.GetCommentById(ctx, 1000)
	if err != nil {
		t.Fatalf("get comment: %v", err)
	}
	if comment.Text.String != "initial text" {
		t.Fatalf("comment text changed to %q", comment.Text.String)
	}
}

func TestAppendCommentInSectionWithActivity_ConcurrencySQLite(t *testing.T) {
	database := openReplyThreadTestDatabaseSQLite(t)
	createReplyThreadTestSchema(t, database)

	execReplySQL(t, database,
		`INSERT INTO users (idusers, username) VALUES (1, 'user1')`,
		`INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (1, 'forum', 'topic', 100, 'append', 1, 'allow')`,
	)

	// Create multiple distinct comments to test concurrent atomic appends.
	// numConcurrentAppends specifies the number of parallel append operations.
	const numConcurrentAppends = 5
	for i := 1; i <= numConcurrentAppends; i++ {
		execReplySQL(t, database,
			fmt.Sprintf(`INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (%d, %d, 1, 'initial text %d', datetime('now', '-5 minutes'))`, 1000+i, 100+i, i),
		)
	}

	queries := NewForDriver(database, "sqlite")
	appender, ok := queries.(CommentAppender)
	if !ok {
		t.Fatalf("queries does not implement CommentAppender")
	}
	ctx := context.Background()

	type appendRes struct {
		commentID int32
		appendID  int64
		err       error
	}
	results := make(chan appendRes, numConcurrentAppends)

	for i := 1; i <= numConcurrentAppends; i++ {
		cID := int32(1000 + i)
		tID := int32(100 + i)
		text := fmt.Sprintf("appended segment %d", i)
		go func() {
			rows, appID, err := appender.AppendCommentInSectionWithActivity(ctx, AppendCommentInSectionForCommenterParams{
				Text:             text,
				AppendWindowMins: 60,
				Section:          "forum",
				ItemType:         sql.NullString{String: "topic", Valid: true},
				GrantUserID:      sql.NullInt32{Int32: 1, Valid: true},
				ItemID:           sql.NullInt32{Int32: 100, Valid: true},
				Written:          sql.NullTime{Time: time.Now().UTC(), Valid: true},
				CommentID:        cID,
				CommenterID:      1,
				ForumthreadID:    tID,
			})
			if err != nil {
				results <- appendRes{commentID: cID, err: err}
				return
			}
			if rows != 1 {
				results <- appendRes{commentID: cID, err: fmt.Errorf("expected 1 row affected, got %d", rows)}
				return
			}
			results <- appendRes{commentID: cID, appendID: appID}
		}()
	}

	seenAppendIDs := make(map[int64]bool)
	for i := 0; i < numConcurrentAppends; i++ {
		select {
		case res := <-results:
			if res.err != nil {
				t.Fatalf("concurrent append error on comment %d: %v", res.commentID, res.err)
			}
			if res.appendID <= 0 {
				t.Fatalf("expected positive append ID, got %d", res.appendID)
			}
			if seenAppendIDs[res.appendID] {
				t.Fatalf("duplicate append ID %d allocated concurrently", res.appendID)
			}
			seenAppendIDs[res.appendID] = true
		case <-time.After(10 * time.Second):
			t.Fatal("timeout waiting for concurrent appends")
		}
	}
}
