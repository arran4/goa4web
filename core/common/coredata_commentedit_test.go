package common_test

import (
	"context"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCommentEditURLsPrivateForum(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) { return 1, nil }
	q.GetForumTopicByIdForUserFn = func(context.Context, db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{Idforumtopic: 30, Handler: "private"}, nil
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 1
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.SetCurrentSection("privateforum")
	cd.SetCurrentThreadAndTopic(106, 30)
	cd.ForumBasePath = "/private"
	cmt := &db.GetCommentsByThreadIdForUserRow{Idcomments: 42, ForumthreadID: 106, UsersIdusers: 1, IsOwner: true}

	if got, want := cd.CommentEditURL(cmt), "?editComment=42#edit"; got != want {
		t.Fatalf("CommentEditURL got %q, want %q", got, want)
	}
	if got, want := cd.CommentEditSaveURL(cmt), "/private/topic/30/thread/106/comment/42"; got != want {
		t.Fatalf("CommentEditSaveURL got %q, want %q", got, want)
	}
}

func TestCommentEditURLsAdminMode(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) { return 1, nil }
	q.GetForumTopicByIdForUserFn = func(context.Context, db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{Idforumtopic: 30, Handler: "forum"}, nil
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 1
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.AdminMode = true
	cd.SetCurrentSection("forum")
	cd.SetCurrentThreadAndTopic(106, 30)
	cmt := &db.GetCommentsByThreadIdForUserRow{Idcomments: 42, ForumthreadID: 106, UsersIdusers: 1, IsOwner: true}

	if got, want := cd.CommentEditURL(cmt), "?editComment=42&mode=admin#edit"; got != want {
		t.Fatalf("CommentEditURL got %q, want %q", got, want)
	}
}

func TestCommentEditSaveURLPrivateForumFallback(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) { return 1, nil }
	q.GetForumTopicByIdForUserFn = func(context.Context, db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{Idforumtopic: 30, Handler: "private"}, nil
	}
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 1
	common.WithUserRoles([]string{"administrator"})(cd)
	cd.SetCurrentSection("privateforum")
	cd.SetCurrentThreadAndTopic(106, 30)
	cmt := &db.GetCommentsByThreadIdForUserRow{Idcomments: 42, ForumthreadID: 106, UsersIdusers: 1, IsOwner: true}

	if got, want := cd.CommentEditSaveURL(cmt), "/forum/topic/30/thread/106/comment/42"; got != want {
		t.Fatalf("CommentEditSaveURL got %q, want %q", got, want)
	}
}

func TestCommentEditURLsNonForum(t *testing.T) {
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) { return 1, nil }
	cd := common.NewTestCoreData(t, q)
	cd.UserID = 1
	cmt := &db.GetCommentsByThreadIdForUserRow{Idcomments: 42, ForumthreadID: 106, UsersIdusers: 1, IsOwner: true}

	cd.SetCurrentSection("blogs")
	if got := cd.CommentEditURL(cmt); !strings.Contains(got, "42") {
		t.Fatalf("CommentEditURL for blogs got %q", got)
	}
	cd.SetCurrentSection("news")
	if got, want := cd.CommentEditURL(cmt), "?editComment=42#edit"; got != want {
		t.Fatalf("CommentEditURL for news got %q, want %q", got, want)
	}
	cd.SetCurrentSection("writing")
	if got, want := cd.CommentEditURL(cmt), "?editComment=42#edit"; got != want {
		t.Fatalf("CommentEditURL for writing got %q, want %q", got, want)
	}
}

func TestCanEditForumCommentActionMatrix(t *testing.T) {
	tests := []struct {
		name       string
		private    bool
		authorID   int32
		grant      string
		adminMode  bool
		adminRoles bool
		want       bool
	}{
		{name: "public author edit", authorID: 1, grant: "edit", want: true},
		{name: "public author edit-any does not replace edit", authorID: 1, grant: "edit-any"},
		{name: "public other edit denied", authorID: 2, grant: "edit"},
		{name: "public other edit-any", authorID: 2, grant: "edit-any", want: true},
		{name: "append-only cannot edit", authorID: 1, grant: "append"},
		{name: "private author thread edit", private: true, authorID: 1, grant: "edit", want: true},
		{name: "private other thread edit-any", private: true, authorID: 2, grant: "edit-any", want: true},
		{name: "admin mode may edit", authorID: 2, adminMode: true, adminRoles: true, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			section := "forum"
			item := "topic"
			currentSection := "forum"
			handler := "forum"
			if tc.private {
				section = "privateforum_thread"
				item = "thread"
				currentSection = "privateforum"
				handler = "private"
			}
			options := []testhelpers.StubOption{}
			if tc.grant != "" {
				options = append(options, testhelpers.WithGrant(section, item, tc.grant))
			}
			q := testhelpers.NewQuerierStub(options...)
			q.GetForumTopicByIdForUserFn = func(context.Context, db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
				return &db.GetForumTopicByIdForUserRow{Idforumtopic: 30, Handler: handler}, nil
			}
			cd := common.NewTestCoreData(t, q)
			cd.UserID = 1
			cd.SetCurrentSection(currentSection)
			cd.SetCurrentThreadAndTopic(106, 30)
			if tc.adminRoles {
				common.WithUserRoles([]string{"administrator"})(cd)
				common.WithPermissions([]*db.GetPermissionsByUserIDRow{{Name: "administrator", IsAdmin: true}})(cd)
			}
			cd.AdminMode = tc.adminMode
			cmt := &db.GetCommentsByThreadIdForUserRow{Idcomments: 42, ForumthreadID: 106, UsersIdusers: tc.authorID, IsOwner: tc.authorID == cd.UserID}
			if got := cd.CanEditComment(cmt); got != tc.want {
				t.Fatalf("CanEditComment() = %v, want %v", got, tc.want)
			}
		})
	}
}
