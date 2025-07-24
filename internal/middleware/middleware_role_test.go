package middleware

import (
	"context"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
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

	mock.ExpectExec("INSERT INTO sessions").WithArgs("sessid", int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	rows := sqlmock.NewRows([]string{"iduser_roles", "users_idusers", "name"}).
		AddRow(1, 1, "moderator")
	mock.ExpectQuery(regexp.QuoteMeta("FROM user_roles")).WithArgs(int32(1)).
		WillReturnRows(rows)

	session := &sessions.Session{ID: "sessid", Values: map[interface{}]interface{}{"UID": int32(1)}}
	req := httptest.NewRequest("GET", "/", nil)
	q := dbpkg.New(db)
	cd := common.NewCoreData(req.Context(), q, common.WithConfig(config.AppRuntimeConfig))
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	var cdOut *common.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdOut, _ = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	})

	CoreAdderMiddlewareWithDB(db, config.AppRuntimeConfig, 0)(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anonymous", "user", "moderator"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
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
	// TODO find a way of avoid tests which impact global state
	mock.MatchExpectationsInOrder(false)

	mock.ExpectExec("DELETE FROM sessions").WithArgs("sessid").
		WillReturnResult(sqlmock.NewResult(0, 0))

	session := &sessions.Session{ID: "sessid"}
	req := httptest.NewRequest("GET", "/", nil)
	q := dbpkg.New(db)
	cd := common.NewCoreData(req.Context(), q, common.WithConfig(config.AppRuntimeConfig))
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	var cdOut *common.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdOut, _ = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	})

	CoreAdderMiddlewareWithDB(db, config.AppRuntimeConfig, 0)(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anonymous"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
