//go:build !sqlite && !sqlite3

package db

import (
	"context"
	"database/sql"
	"testing"
)

func testUpdateEligibilityMatrix(t *testing.T, database *sql.DB, queries Querier, ctx context.Context) {
	mustExec(t, ctx, database, "INSERT INTO users (idusers, username) VALUES (?, ?)", 60, "poster")
	mustExec(t, ctx, database, "INSERT INTO users (idusers, username) VALUES (?, ?)", 61, "editor")
	mustExec(t, ctx, database, "INSERT INTO roles (id, name, is_admin) VALUES (?, ?, ?)", 101, "administrator", 1)

	mustExec(t, ctx, database, "INSERT INTO forumtopic (idforumtopic, handler) VALUES (?, ?)", 30, "forum")
	mustExec(t, ctx, database, "INSERT INTO forumthread (idforumthread, firstpost, lastposter, forumtopic_idforumtopic) VALUES (?, ?, ?, ?)", 1000, 2000, 60, 30)

	tests := []struct {
		name     string
		setup    func()
		editorID int32
		wantRows bool
	}{
		{
			name: "owner + edit -> 1 row",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "forum", "topic", 30, "edit", 1, "allow")
			},
			editorID: 60,
			wantRows: true,
		},
		{
			name:     "owner + no edit -> 0 rows",
			setup:    func() {},
			editorID: 60,
			wantRows: false,
		},
		{
			name: "other user + edit -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 61, "forum", "topic", 30, "edit", 1, "allow")
			},
			editorID: 61,
			wantRows: false,
		},
		{
			name: "other user + edit-any -> 1 row",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 61, "forum", "topic", 30, "edit-any", 1, "allow")
			},
			editorID: 61,
			wantRows: true,
		},
		{
			name: "append-only -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "forum", "topic", 30, "append", 1, "allow")
			},
			editorID: 60,
			wantRows: false,
		},
		{
			name: "edit-any via editor role -> 1 row",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO grants (role_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 100, "forum", "topic", 30, "edit-any", 1, "allow")
				mustExec(t, ctx, database, "INSERT INTO user_roles (users_idusers, role_id) VALUES (?, ?)", 61, 100)
			},
			editorID: 61,
			wantRows: true,
		},
		{
			name: "administrator may edit another user's comment",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO user_roles (users_idusers, role_id) VALUES (?, ?)", 61, 101)
			},
			editorID: 61,
			wantRows: true,
		},
		{
			name: "public thread + only privateforum_thread edit grant -> 0 rows",
			setup: func() {
				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "privateforum_thread", "thread", 1000, "edit", 1, "allow")
			},
			editorID: 60,
			wantRows: false,
		},
		{
			name: "private thread + only forum/topic edit grant -> 0 rows",
			setup: func() {
				// Make topic private
				mustExec(t, ctx, database, "UPDATE forumtopic SET handler = 'private' WHERE idforumtopic = 30")
				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "forum", "topic", 30, "edit", 1, "allow")
			},
			editorID: 60,
			wantRows: false,
		},
		{
			name: "private forum coverage: owner + edit -> 1 row",
			setup: func() {
				mustExec(t, ctx, database, "UPDATE forumtopic SET handler = 'private' WHERE idforumtopic = 30")
				mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "privateforum_thread", "thread", 1000, "edit", 1, "allow")
			},
			editorID: 60,
			wantRows: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mustExec(t, ctx, database, "DELETE FROM comments")
			mustExec(t, ctx, database, "DELETE FROM grants")
			mustExec(t, ctx, database, "DELETE FROM user_roles")
			mustExec(t, ctx, database, "UPDATE forumtopic SET handler = 'forum' WHERE idforumtopic = 30")

			mustExec(t, ctx, database, "INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (?, ?, ?, ?, '2030-01-01 12:00:00')", 2000, 1000, 60, "Initial post")

			tc.setup()

			rowsAffected, err := queries.UpdateForumCommentForEditor(ctx, UpdateForumCommentForEditorParams{
				Text:         sql.NullString{String: "updated text", Valid: true},
				CommentID:    2000,
				EditorID:     tc.editorID,
				EditorUserID: sql.NullInt32{Int32: tc.editorID, Valid: true},
			})
			if err != nil {
				t.Fatalf("UpdateForumCommentForEditor: %v", err)
			}

			hasUpdated := rowsAffected > 0
			if hasUpdated != tc.wantRows {
				t.Fatalf("Expected wantRows=%v, got=%v", tc.wantRows, hasUpdated)
			}
			if tc.wantRows {
				cRow, err := queries.GetCommentById(ctx, 2000)
				if err != nil {
					t.Fatalf("GetCommentById: %v", err)
				}
				if cRow.Text.String != "updated text" {
					t.Fatalf("Expected updated text, got: %q", cRow.Text.String)
				}
			}
		})
	}

	t.Run("generic non-forum self-edit semantics remain unchanged", func(t *testing.T) {
		mustExec(t, ctx, database, "DELETE FROM grants")
		mustExec(t, ctx, database, "DELETE FROM comments")
		mustExec(t, ctx, database, "INSERT INTO comments (idcomments, forumthread_id, users_idusers, text, written) VALUES (?, ?, ?, ?, '2030-01-01 12:00:00')", 2000, 1000, 60, "Initial post")
		mustExec(t, ctx, database, "INSERT INTO grants (user_id, section, item, item_id, action, active, rule_type) VALUES (?, ?, ?, ?, ?, ?, ?)", 60, "forum", "comment", 2000, "edit", 1, "allow")
		rows, err := queries.UpdateCommentForEditor(ctx, UpdateCommentForEditorParams{
			Text: sql.NullString{String: "owner update", Valid: true}, CommentID: 2000,
			CommenterID: 60, EditorID: sql.NullInt32{Int32: 60, Valid: true},
		})
		if err != nil || rows != 1 {
			t.Fatalf("owner generic update rows=%d err=%v", rows, err)
		}
		rows, err = queries.UpdateCommentForEditor(ctx, UpdateCommentForEditorParams{
			Text: sql.NullString{String: "other update", Valid: true}, CommentID: 2000,
			CommenterID: 61, EditorID: sql.NullInt32{Int32: 61, Valid: true},
		})
		if err != nil || rows != 0 {
			t.Fatalf("other-user generic update rows=%d err=%v", rows, err)
		}
	})
}

func TestMySQLUpdateEligibilityMatrix(t *testing.T) {
	database := openReplyThreadTestDatabase(t)
	createReplyThreadTestSchema(t, database)
	queries := NewForDriver(database, "mysql")
	ctx := context.Background()

	mustExec(t, ctx, database, "DELETE FROM grants")
	mustExec(t, ctx, database, "DELETE FROM content_read_markers")
	mustExec(t, ctx, database, "DELETE FROM comments")
	mustExec(t, ctx, database, "DELETE FROM forumthread")
	mustExec(t, ctx, database, "DELETE FROM users")
	mustExec(t, ctx, database, "DELETE FROM user_roles")
	mustExec(t, ctx, database, "UPDATE forumtopic SET handler = 'forum' WHERE idforumtopic = 30")

	testUpdateEligibilityMatrix(t, database, queries, ctx)
}
