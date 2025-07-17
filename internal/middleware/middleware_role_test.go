package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	corecommon "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/sessions"
)

func TestCoreAdderMiddlewareUserRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	q := dbpkg.New(db)

	mock.ExpectExec("INSERT INTO sessions").WithArgs("sessid", int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	rows := sqlmock.NewRows([]string{"iduser_roles", "users_idusers", "name"}).
		AddRow(1, 1, "moderator")
	mock.ExpectQuery(regexp.QuoteMeta("FROM user_roles")).WithArgs(int32(1)).
		WillReturnRows(rows)

	session := &sessions.Session{ID: "sessid", Values: map[interface{}]interface{}{"UID": int32(1)}}
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), handlers.KeyQueries, q)
	ctx = context.WithValue(ctx, corecommon.ContextValues("session"), session)
	req = req.WithContext(ctx)

	var cd *corecommon.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, _ = r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData)
	})

	CoreAdderMiddleware(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anonymous", "user", "moderator"}
	if diff := cmp.Diff(want, cd.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCoreAdderMiddlewareAnonymous(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	q := dbpkg.New(db)

	mock.ExpectExec("DELETE FROM sessions").WithArgs("sessid").
		WillReturnResult(sqlmock.NewResult(0, 0))

	session := &sessions.Session{ID: "sessid"}
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), handlers.KeyQueries, q)
	ctx = context.WithValue(ctx, corecommon.ContextValues("session"), session)
	req = req.WithContext(ctx)

	var cd *corecommon.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd, _ = r.Context().Value(handlers.KeyCoreData).(*corecommon.CoreData)
	})

	CoreAdderMiddleware(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anonymous"}
	if diff := cmp.Diff(want, cd.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
