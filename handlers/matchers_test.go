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
	q := &db.QuerierStub{
		SystemCheckGrantStubs: []db.SystemCheckGrantStub{{Result: 1}},
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "1"}}
	if !RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to allow request")
	}
	if got := len(q.SystemCheckGrantCalls); got != 1 {
		t.Fatalf("expected 1 grant check, got %d", got)
	}
	call := q.SystemCheckGrantCalls[0]
	if call.Section != "news" || call.Action != "edit" {
		t.Fatalf("unexpected grant check: %+v", call)
	}
	if call.ItemID.Int32 != 1 || !call.ItemID.Valid {
		t.Fatalf("expected item ID 1, got %+v", call.ItemID)
	}
	if call.UserID.Int32 != 1 || !call.UserID.Valid {
		t.Fatalf("expected user ID 1, got %+v", call.UserID)
	}
}

func TestRequireGrantDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/news/2/edit", nil)
	q := &db.QuerierStub{
		SystemCheckGrantStubs: []db.SystemCheckGrantStub{{Err: sql.ErrNoRows}},
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig())
	cd.UserID = 2
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "2"}}
	if RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to reject request")
	}
	if got := len(q.SystemCheckGrantCalls); got != 1 {
		t.Fatalf("expected 1 grant check, got %d", got)
	}
	call := q.SystemCheckGrantCalls[0]
	if call.ItemID.Int32 != 2 || !call.ItemID.Valid {
		t.Fatalf("expected item ID 2, got %+v", call.ItemID)
	}
	if call.UserID.Int32 != 2 || !call.UserID.Valid {
		t.Fatalf("expected user ID 2, got %+v", call.UserID)
	}
}
