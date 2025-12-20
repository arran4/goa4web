package common_test

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/db/testutil"
)

func TestCommentEditURLsPrivateForum(t *testing.T) {
	queries := testutil.NewBaseQuerier(t)
	queries.AllowGrants()
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
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
	queries := testutil.NewBaseQuerier(t)
	queries.AllowGrants()
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.SetCurrentSection("privateforum")
	cd.SetCurrentThreadAndTopic(106, 30)
	cmt := &db.GetCommentsByThreadIdForUserRow{Idcomments: 42, IsOwner: true}

	if got, want := cd.CommentEditSaveURL(cmt), "/forum/topic/30/thread/106/comment/42"; got != want {
		t.Fatalf("CommentEditSaveURL got %q, want %q", got, want)
	}
}
