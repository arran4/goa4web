package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/sessions"
)

func TestCoreDataMiddlewareUserRoles(t *testing.T) {
	navReg := nav.NewRegistry()

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	cfg := config.NewRuntimeConfig()
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	mock.ExpectExec("INSERT INTO sessions").WithArgs("sessid", int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	rows := sqlmock.NewRows([]string{"iduser_roles", "users_idusers", "role_id", "name", "is_admin"}).
		AddRow(1, 1, 2, "moderator", false)
	mock.ExpectQuery(regexp.QuoteMeta("FROM user_roles")).WithArgs(int32(1)).
		WillReturnRows(rows)

	session := &sessions.Session{ID: "sessid", Values: map[interface{}]interface{}{"UID": int32(1)}}
	req := httptest.NewRequest("GET", "/", nil)
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	var cdOut *common.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdOut, _ = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	})

	reg := email.NewRegistry()
	signer := "k"
	linkSigner := "k"
	srv := New(
		WithDB(conn),
		WithConfig(cfg),
		WithEmailRegistry(reg),
		WithImageSignKey(signer),
		WithLinkSignKey(linkSigner),
		WithNavRegistry(navReg),
	)
	srv.CoreDataMiddleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anyone", "user", "moderator"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCoreDataMiddlewareAnonymous(t *testing.T) {
	navReg := nav.NewRegistry()

	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	cfg := config.NewRuntimeConfig()
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	mock.ExpectExec("DELETE FROM sessions").WithArgs("sessid").
		WillReturnResult(sqlmock.NewResult(0, 0))

	session := &sessions.Session{ID: "sessid"}
	req := httptest.NewRequest("GET", "/", nil)
	q := db.New(conn)
	cd := common.NewCoreData(req.Context(), q, cfg)
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	var cdOut *common.CoreData
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdOut, _ = r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	})

	reg := email.NewRegistry()
	signer := "k"
	linkSigner := "k"
	srv := New(
		WithDB(conn),
		WithConfig(cfg),
		WithEmailRegistry(reg),
		WithImageSignKey(signer),
		WithLinkSignKey(linkSigner),
		WithNavRegistry(navReg),
	)
	srv.CoreDataMiddleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	want := []string{"anyone"}
	if diff := cmp.Diff(want, cdOut.UserRoles()); diff != "" {
		t.Fatalf("roles mismatch (-want +got):\n%s", diff)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
