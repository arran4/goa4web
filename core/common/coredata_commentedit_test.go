package common_test

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCommentEditURLsPrivateForum(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}

	cd := common.NewTestCoreData(t, q)
	cd.UserID = 1
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.SetCurrentSection("privateforum")
	cd.SetCurrentThreadAndTopic(106, 30)
	cd.ForumBasePath = "/private"
	q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{Idforumtopic: 30, Handler: "private"}, nil
	}

	cmt := &db.GetCommentsByThreadIdForUserRow{
		Idcomments:    42,
		ForumthreadID: 106,
		UsersIdusers:  1,
		IsOwner:       true,
	}

	if got, want := cd.CommentEditURL(cmt), "?editComment=42#edit"; got != want {
		t.Fatalf("CommentEditURL got %q, want %q", got, want)
	}
	if got, want := cd.CommentEditSaveURL(cmt), "/private/topic/30/thread/106/comment/42"; got != want {
		t.Fatalf("CommentEditSaveURL got %q, want %q", got, want)
	}
}

func TestCommentEditURLsAdminMode(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}

	cd := common.NewTestCoreData(t, q)
	cd.UserID = 1
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.AdminMode = true
	cd.SetCurrentSection("forum")
	q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{Idforumtopic: 30, Handler: "forum"}, nil
	}

	cd.SetCurrentThreadAndTopic(106, 30)

	cmt := &db.GetCommentsByThreadIdForUserRow{
		Idcomments:    42,
		ForumthreadID: 106,
		UsersIdusers:  1,
		IsOwner:       true,
	}

	got := cd.CommentEditURL(cmt)
	// url.Values.Encode sorts by key: editComment comes before mode
	want := "?editComment=42&mode=admin#edit"
	if got != want {
		t.Fatalf("CommentEditURL got %q, want %q", got, want)
	}
}

func TestCommentEditSaveURLPrivateForumFallback(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}

	cd := common.NewTestCoreData(t, q)
	cd.UserID = 1
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.SetCurrentSection("privateforum")
	cd.SetCurrentThreadAndTopic(106, 30)
	q.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{Idforumtopic: 30, Handler: "private"}, nil
	}

	cmt := &db.GetCommentsByThreadIdForUserRow{
		Idcomments:    42,
		ForumthreadID: 106,
		UsersIdusers:  1,
		IsOwner:       true,
	}

	if got, want := cd.CommentEditSaveURL(cmt), "/forum/topic/30/thread/106/comment/42"; got != want {
		t.Fatalf("CommentEditSaveURL got %q, want %q", got, want)
	}
}
