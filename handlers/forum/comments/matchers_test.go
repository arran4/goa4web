package comments

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestRequireCommentAuthor_AllowsAuthor(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	commentID := int32(3)
	threadID := int32(5)
	userID := int32(7)

	mock.ExpectQuery("(?s)WITH role_ids AS .*FROM comments c").
		WithArgs(userID, userID, commentID, userID, userID, sql.NullInt32{Int32: userID, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idcomments", "forumthread_id", "users_idusers", "language_id", "written", "text", "timezone", "deleted_at", "last_index", "username", "is_owner",
		}).AddRow(commentID, threadID, userID, sql.NullInt32{}, sql.NullTime{}, sql.NullString{}, sql.NullString{}, sql.NullTime{}, sql.NullTime{}, sql.NullString{}, true))

	req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/thread/5/comment/3", nil)
	req = mux.SetURLVars(req, map[string]string{"comment": "3"})

	sess := &sessions.Session{Values: map[interface{}]interface{}{"UID": userID}}
	q := db.New(conn)
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRequireCommentAuthor_AllowsGrantHolder(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	commentID := int32(9)
	threadID := int32(10)
	authorID := int32(11)
	adminID := int32(12)

	mock.ExpectQuery("(?s)WITH role_ids AS .*FROM comments c").
		WithArgs(adminID, adminID, commentID, adminID, adminID, sql.NullInt32{Int32: adminID, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idcomments", "forumthread_id", "users_idusers", "language_id", "written", "text", "timezone", "deleted_at", "last_index", "username", "is_owner",
		}).AddRow(commentID, threadID, authorID, sql.NullInt32{}, sql.NullTime{}, sql.NullString{}, sql.NullString{}, sql.NullTime{}, sql.NullTime{}, sql.NullString{}, false))

	mock.ExpectQuery("(?s).*SELECT 1 FROM grants.*").
		WithArgs(adminID, "forum", sql.NullString{String: "thread", Valid: true}, "edit-any", sql.NullInt32{Int32: threadID, Valid: true}, sql.NullInt32{Int32: adminID, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/thread/10/comment/9", nil)
	req = mux.SetURLVars(req, map[string]string{"comment": "9"})

	sess := &sessions.Session{Values: map[interface{}]interface{}{"UID": adminID}}
	q := db.New(conn)
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRequireCommentAuthor_AllowsAdminMode(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	commentID := int32(13)
	threadID := int32(15)
	authorID := int32(16)
	adminID := int32(17)

	mock.ExpectQuery("(?s)WITH role_ids AS .*FROM comments c").
		WithArgs(adminID, adminID, commentID, adminID, adminID, sql.NullInt32{Int32: adminID, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{
			"idcomments", "forumthread_id", "users_idusers", "language_id", "written", "text", "timezone", "deleted_at", "last_index", "username", "is_owner",
		}).AddRow(commentID, threadID, authorID, sql.NullInt32{}, sql.NullTime{}, sql.NullString{}, sql.NullString{}, sql.NullTime{}, sql.NullTime{}, sql.NullString{}, false))

	// HasAdminRole call
	mock.ExpectQuery("SELECT .* FROM user_roles .* JOIN roles .* WHERE .*is_admin = 1").
		WithArgs(int32(adminID)).
		WillReturnRows(sqlmock.NewRows([]string{"iduser_roles", "users_idusers", "role_id"}).AddRow(1, adminID, 1))

	req := httptest.NewRequest(http.MethodPost, "/forum/topic/1/thread/15/comment/13", nil)
	req = mux.SetURLVars(req, map[string]string{"comment": "13"})

	sess := &sessions.Session{Values: map[interface{}]interface{}{"UID": adminID}}
	q := db.New(conn)
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

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
