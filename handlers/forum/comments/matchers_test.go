package comments

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestRequireCommentAuthor_AllowsAuthor(t *testing.T) {
	commentID := int32(3)
	threadID := int32(5)
	userID := int32(7)

	q := &db.QuerierStub{
		GetCommentByIdForUserRow: &db.GetCommentByIdForUserRow{
			Idcomments:    commentID,
			ForumthreadID: threadID,
			UsersIdusers:  userID,
			IsOwner:       true,
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/thread/5/comment/3", nil)
	req = mux.SetURLVars(req, map[string]string{"comment": "3"})

	sess := &sessions.Session{Values: map[interface{}]interface{}{"UID": userID}}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithSession(sess), common.WithUserRoles([]string{"anyone", "user"}))

	ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	RequireCommentAuthor(handler).ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected downstream handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rr.Code)
	}

	if len(q.GetCommentByIdForUserCalls) != 1 {
		t.Fatalf("expected one comment lookup, got %d", len(q.GetCommentByIdForUserCalls))
	}
	want := db.GetCommentByIdForUserParams{
		ViewerID: userID,
		ID:       commentID,
		UserID:   sql.NullInt32{Int32: userID, Valid: true},
	}
	if got := q.GetCommentByIdForUserCalls[0]; got != want {
		t.Fatalf("unexpected comment lookup params: %#v", got)
	}
	if len(q.SystemCheckGrantCalls) != 0 {
		t.Fatalf("unexpected grant checks: %v", q.SystemCheckGrantCalls)
	}
}

func TestRequireCommentAuthor_AllowsGrantHolder(t *testing.T) {
	commentID := int32(9)
	threadID := int32(10)
	authorID := int32(11)
	adminID := int32(12)

	q := &db.QuerierStub{
		GetCommentByIdForUserRow: &db.GetCommentByIdForUserRow{
			Idcomments:    commentID,
			ForumthreadID: threadID,
			UsersIdusers:  authorID,
			IsOwner:       false,
		},
		SystemCheckGrantFn: func(arg db.SystemCheckGrantParams) (int32, error) {
			if arg.Action == "edit-any" {
				return 1, nil
			}
			return 0, sql.ErrNoRows
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/thread/10/comment/9", nil)
	req = mux.SetURLVars(req, map[string]string{"comment": "9"})

	sess := &sessions.Session{Values: map[interface{}]interface{}{"UID": adminID}}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithSession(sess), common.WithUserRoles([]string{"anyone", "user"}))

	ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	RequireCommentAuthor(handler).ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected downstream handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rr.Code)
	}

	if len(q.GetCommentByIdForUserCalls) != 1 {
		t.Fatalf("expected one comment lookup, got %d", len(q.GetCommentByIdForUserCalls))
	}
	want := db.GetCommentByIdForUserParams{
		ViewerID: adminID,
		ID:       commentID,
		UserID:   sql.NullInt32{Int32: adminID, Valid: true},
	}
	if got := q.GetCommentByIdForUserCalls[0]; got != want {
		t.Fatalf("unexpected comment lookup params: %#v", got)
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one grant check, got %d", len(q.SystemCheckGrantCalls))
	}
	if got := q.SystemCheckGrantCalls[0]; got.Action != "edit-any" || got.Section != "forum" || got.Item != (sql.NullString{String: "thread", Valid: true}) || got.ItemID != (sql.NullInt32{Int32: threadID, Valid: true}) {
		t.Fatalf("unexpected grant params: %#v", got)
	}
}

func TestRequireCommentAuthor_AllowsAdminMode(t *testing.T) {
	commentID := int32(13)
	threadID := int32(15)
	authorID := int32(16)
	adminID := int32(17)

	q := &db.QuerierStub{
		GetCommentByIdForUserRow: &db.GetCommentByIdForUserRow{
			Idcomments:    commentID,
			ForumthreadID: threadID,
			UsersIdusers:  authorID,
			IsOwner:       false,
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/thread/15/comment/13", nil)
	req = mux.SetURLVars(req, map[string]string{"comment": "13"})

	sess := &sessions.Session{Values: map[interface{}]interface{}{"UID": adminID}}
	cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig(), common.WithSession(sess), common.WithUserRoles([]string{"anyone", "user", "administrator"}))
	cd.AdminMode = true

	ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	called := false
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	RequireCommentAuthor(handler).ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected downstream handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rr.Code)
	}

	if len(q.GetCommentByIdForUserCalls) != 1 {
		t.Fatalf("expected one comment lookup, got %d", len(q.GetCommentByIdForUserCalls))
	}
	want := db.GetCommentByIdForUserParams{
		ViewerID: adminID,
		ID:       commentID,
		UserID:   sql.NullInt32{Int32: adminID, Valid: true},
	}
	if got := q.GetCommentByIdForUserCalls[0]; got != want {
		t.Fatalf("unexpected comment lookup params: %#v", got)
	}
	if len(q.SystemCheckGrantCalls) != 0 {
		t.Fatalf("unexpected grant checks: %v", q.SystemCheckGrantCalls)
	}
}
