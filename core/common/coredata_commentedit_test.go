package common_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestCommentEditURLsPrivateForum(t *testing.T) {
	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.SetCurrentSection("privateforum")
	cd.SetCurrentThreadAndTopic(106, 30)
	cd.ForumBasePath = "/private"
	cmt := &db.GetCommentsByThreadIdForUserRow{Idcomments: 42, IsOwner: true}

	if got, want := cd.CommentEditURL(cmt), "?comment=42#edit"; got != want {
		t.Fatalf("CommentEditURL got %q, want %q", got, want)
	}
	if got, want := cd.CommentEditSaveURL(cmt), "/private/topic/30/thread/106/comment/42"; got != want {
		t.Fatalf("CommentEditSaveURL got %q, want %q", got, want)
	}
}

func TestCommentEditSaveURLPrivateForumFallback(t *testing.T) {
	conn, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.SetCurrentSection("privateforum")
	cd.SetCurrentThreadAndTopic(106, 30)
	cmt := &db.GetCommentsByThreadIdForUserRow{Idcomments: 42, IsOwner: true}

	if got, want := cd.CommentEditSaveURL(cmt), "/forum/topic/30/thread/106/comment/42"; got != want {
		t.Fatalf("CommentEditSaveURL got %q, want %q", got, want)
	}
}
