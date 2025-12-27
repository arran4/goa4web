package handlers

import (
	"context"
	"errors"
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestRequiredGrantAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	q := &db.QuerierStub{}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if !RequiredGrant("blogs", "entry", "post", 0)(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}
}

func TestRequiredGrantDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	q := &db.QuerierStub{SystemCheckGrantErr: errors.New("denied")}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if RequiredGrant("blogs", "entry", "post", 0)(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}
}

func TestRequireGrantAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/news/1/edit", nil)
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.ExpectQuery(regexp.QuoteMeta("WITH role_ids AS (\n    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = ?\n)\nSELECT 1 FROM grants g\nWHERE g.section = ?\n  AND (g.item = ? OR g.item IS NULL)\n  AND g.action = ?\n  AND g.active = 1\n  AND (g.item_id = ? OR g.item_id IS NULL)\n  AND (g.user_id = ? OR g.user_id IS NULL)\n  AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))\nLIMIT 1\n")).
		WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "edit", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	cd := common.NewCoreData(req.Context(), db.New(conn), config.NewRuntimeConfig())
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "1"}}
	if !RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to allow request")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("grant expectations: %v", err)
	}
}

func TestRequireGrantDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/news/2/edit", nil)
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.ExpectQuery(regexp.QuoteMeta("WITH role_ids AS (\n    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = ?\n)\nSELECT 1 FROM grants g\nWHERE g.section = ?\n  AND (g.item = ? OR g.item IS NULL)\n  AND g.action = ?\n  AND g.active = 1\n  AND (g.item_id = ? OR g.item_id IS NULL)\n  AND (g.user_id = ? OR g.user_id IS NULL)\n  AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))\nLIMIT 1\n")).
		WithArgs(int32(2), "news", sql.NullString{String: "post", Valid: true}, "edit", sql.NullInt32{Int32: 2, Valid: true}, sql.NullInt32{Int32: 2, Valid: true}).
		WillReturnError(sql.ErrNoRows)

	cd := common.NewCoreData(req.Context(), db.New(conn), config.NewRuntimeConfig())
	cd.UserID = 2
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "2"}}
	if RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to reject request")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("grant expectations: %v", err)
	}
}
