package forum

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/handlers/handlertest"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

// B. Test performForumReply with QuerierStub
func TestPerformForumReply_SuccessfulAppend(t *testing.T) {
	stub := testhelpers.NewQuerierStub()
	cd := common.NewCoreData(context.Background(), stub, config.NewRuntimeConfig(), common.WithUserRoles([]string{"member"}))

	// Mock comments lookup
	stub.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		return &db.GetCommentByIdForUserRow{Idcomments: 200, Text: sql.NullString{String: "old text", Valid: true}}, nil
	}

	stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 200}}, nil
	}

	appended := false
	stub.AppendCommentInSectionForCommenterFn = func(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
		appended = true
		return 1, nil // 1 row affected
	}

	stub.GetCommentByIdFn = func(ctx context.Context, idcomments int32) (*db.Comment, error) {
		return &db.Comment{Idcomments: 200, Text: sql.NullString{String: "canonical", Valid: true}}, nil
	}

	stub.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		t.Errorf("CreateCommentInSectionForCommenter should not be called")
		return 300, nil
	}

	thread := &db.GetThreadLastPosterAndPermsForUserRow{Idforumthread: 412}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: 37, Handler: "forum"}

	// We don't need a full Config for the append wrapper inside performForumReply,
	// wait, AttemptAppendForumComment checks cd.Config.

	res, err := performForumReply(cd, 1, thread, topic, 1, "test")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if !appended {
		t.Errorf("expected append to be called")
	}
	if !res.Appended {
		t.Errorf("expected Appended=true")
	}
	if res.CommentID != 200 {
		t.Errorf("expected CommentID=200, got %d", res.CommentID)
	}
	if !res.CanonicalTextOK {
		t.Errorf("expected CanonicalTextOK=true")
	}
	if res.CommentText != "canonical" {
		t.Errorf("expected canonical text")
	}
}

func TestPerformForumReply_ZeroRowsFallback(t *testing.T) {
	stub := testhelpers.NewQuerierStub()
	cd := common.NewCoreData(context.Background(), stub, config.NewRuntimeConfig(), common.WithUserRoles([]string{"member"}))

	stub.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		return &db.GetCommentByIdForUserRow{Idcomments: 200, Text: sql.NullString{String: "old text", Valid: true}}, nil
	}

	stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 200}}, nil
	}

	appended := false
	stub.AppendCommentInSectionForCommenterFn = func(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
		appended = true
		return 0, nil // 0 rows affected -> fallback
	}

	created := false
	stub.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		created = true
		return 300, nil
	}

	thread := &db.GetThreadLastPosterAndPermsForUserRow{Idforumthread: 412}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: 37, Handler: "forum"}

	res, err := performForumReply(cd, 1, thread, topic, 1, "test")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if !appended {
		t.Errorf("expected append to be attempted")
	}
	if !created {
		t.Errorf("expected create to be called")
	}
	if res.Appended {
		t.Errorf("expected Appended=false")
	}
	if res.CommentID != 300 {
		t.Errorf("expected CommentID=300")
	}
}

func TestPerformForumReply_AppendDBError(t *testing.T) {
	stub := testhelpers.NewQuerierStub()
	cd := common.NewCoreData(context.Background(), stub, config.NewRuntimeConfig(), common.WithUserRoles([]string{"member"}))

	stub.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		return &db.GetCommentByIdForUserRow{Idcomments: 200, Text: sql.NullString{String: "old text", Valid: true}}, nil
	}

	stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 200}}, nil
	}

	stub.AppendCommentInSectionForCommenterFn = func(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
		return 0, fmt.Errorf("db error")
	}

	stub.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		t.Errorf("CreateCommentInSectionForCommenter should not be called")
		return 300, nil
	}

	thread := &db.GetThreadLastPosterAndPermsForUserRow{Idforumthread: 412}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: 37, Handler: "forum"}

	_, err := performForumReply(cd, 1, thread, topic, 1, "test")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPerformForumReply_ThreadCommentsError(t *testing.T) {
	stub := testhelpers.NewQuerierStub()
	cd := common.NewCoreData(context.Background(), stub, config.NewRuntimeConfig(), common.WithUserRoles([]string{"member"}))

	stub.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		return &db.GetCommentByIdForUserRow{Idcomments: 200, Text: sql.NullString{String: "old text", Valid: true}}, nil
	}

	stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return nil, fmt.Errorf("db error")
	}

	stub.AppendCommentInSectionForCommenterFn = func(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
		t.Errorf("AppendCommentInSectionForCommenter should not be called")
		return 0, nil
	}

	stub.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		t.Errorf("CreateCommentInSectionForCommenter should not be called")
		return 300, nil
	}

	thread := &db.GetThreadLastPosterAndPermsForUserRow{Idforumthread: 412}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: 37, Handler: "forum"}

	_, err := performForumReply(cd, 1, thread, topic, 1, "test")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPerformForumReply_ReloadFailed(t *testing.T) {
	stub := testhelpers.NewQuerierStub()
	cd := common.NewCoreData(context.Background(), stub, config.NewRuntimeConfig(), common.WithUserRoles([]string{"member"}))

	stub.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		return &db.GetCommentByIdForUserRow{Idcomments: 200, Text: sql.NullString{String: "old text", Valid: true}}, nil
	}

	stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 200}}, nil
	}

	stub.AppendCommentInSectionForCommenterFn = func(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
		return 1, nil // 1 row affected
	}

	stub.GetCommentByIdFn = func(ctx context.Context, idcomments int32) (*db.Comment, error) {
		return nil, fmt.Errorf("reload failed")
	}

	stub.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		t.Errorf("CreateCommentInSectionForCommenter should not be called")
		return 300, nil
	}

	thread := &db.GetThreadLastPosterAndPermsForUserRow{Idforumthread: 412}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: 37, Handler: "forum"}

	res, err := performForumReply(cd, 1, thread, topic, 1, "test")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if !res.Appended {
		t.Errorf("expected Appended=true")
	}
	if res.CanonicalTextOK {
		t.Errorf("expected CanonicalTextOK=false")
	}
	if res.CommentID != 200 {
		t.Errorf("expected CommentID=200, got %d", res.CommentID)
	}
}

// C. Thin HTTP test for ReplyTask
func TestReplyTaskAction_Redirect(t *testing.T) {
	req := httptest.NewRequest("POST", "/forum/topic/37/thread/412", nil)
	session := &sessions.Session{Values: make(map[any]any)}
	if session.Values == nil {
		session.Values = make(map[any]any)
	}
	session.Values["UID"] = int32(1)
	req, cd, stub := handlertest.RequestWithCoreData(t, req, common.WithUserRoles([]string{"member"}), common.WithSession(session))
	cd.UserID = 1
	req = mux.SetURLVars(req, map[string]string{
		"topic":  "37",
		"thread": "412",
	})

	form := url.Values{}
	form.Set("replytext", "reply content")
	form.Set("task", string(TaskReply))
	req.Body = io.NopCloser(strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	stub.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{Idforumtopic: 37, Handler: "forum"}, nil
	}
	stub.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
		return &db.GetThreadLastPosterAndPermsForUserRow{Idforumthread: 412}, nil
	}

	// Fallback path
	stub.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		return []*db.GetCommentsByThreadIdForUserRow{{Idcomments: 200}}, nil
	}
	stub.AppendCommentInSectionForCommenterFn = func(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
		return 0, nil // zero rows -> fallback
	}
	stub.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		return 300, nil
	}

	stub.GetCommentByIdFn = func(ctx context.Context, idcomments int32) (*db.Comment, error) {
		return &db.Comment{Idcomments: idcomments}, nil
	}
	stub.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		return &db.GetCommentByIdForUserRow{Idcomments: 200}, nil
	}

	stub.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}

	cd.ForumBasePath = "/forum"

	rr := httptest.NewRecorder()
	ret := ReplyTaskHandler.Action(rr, req)
	if err, ok := ret.(error); ok {
		t.Fatalf("Action returned error: %v", err)
	}

	typeName := fmt.Sprintf("%T", ret)
	if strings.Contains(typeName, "RedirectHandler") {
		http.Redirect(rr, req, fmt.Sprintf("%s", ret), http.StatusFound)
	} else {
		t.Fatalf("Action returned unhandled type: %s", typeName)
	}

	res := rr.Result()
	if res.StatusCode != http.StatusFound {
		t.Errorf("expected 302 Found, got %d", res.StatusCode)
	}
	if loc := res.Header.Get("Location"); !strings.Contains(loc, "/forum/topic/37/thread/412#c1") {
		t.Errorf("expected redirect to #c300, got %s", loc)
	}
}
