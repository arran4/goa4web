package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestRequiredGrantAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if !RequiredGrant("blogs", "entry", "post", 0)(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}
}

func TestRequiredGrantDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 0, errors.New("denied")
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if RequiredGrant("blogs", "entry", "post", 0)(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}
}

func TestRequireGrantAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/news/1/edit", nil)
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 1, nil
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "1"}}
	if !RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to allow request")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one grant check, got %d", len(q.SystemCheckGrantCalls))
	}
	want := db.SystemCheckGrantParams{
		ViewerID: cd.UserID,
		Section:  "news",
		Item:     sql.NullString{String: "post", Valid: true},
		Action:   "edit",
		ItemID:   sql.NullInt32{Int32: 1, Valid: true},
		UserID:   sql.NullInt32{Int32: 1, Valid: true},
	}
	if got := q.SystemCheckGrantCalls[0]; got != want {
		t.Fatalf("unexpected grant check: %#v", got)
	}
}

func TestRequireGrantDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/news/2/edit", nil)
	q := testhelpers.NewQuerierStub()
	q.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 0, sql.ErrNoRows
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	cd.UserID = 2
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "2"}}
	if RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to reject request")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected one grant check, got %d", len(q.SystemCheckGrantCalls))
	}
	want := db.SystemCheckGrantParams{
		ViewerID: cd.UserID,
		Section:  "news",
		Item:     sql.NullString{String: "post", Valid: true},
		Action:   "edit",
		ItemID:   sql.NullInt32{Int32: 2, Valid: true},
		UserID:   sql.NullInt32{Int32: 2, Valid: true},
	}
	if got := q.SystemCheckGrantCalls[0]; got != want {
		t.Fatalf("unexpected grant check: %#v", got)
	}
}
