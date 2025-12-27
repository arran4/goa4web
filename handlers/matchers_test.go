package handlers

import (
	"context"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestRequiredGrantAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	q := &db.QuerierStub{}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if !RequiredGrant("blogs", "entry", "post", 0)(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}
}

func TestRequiredGrantDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	q := &db.QuerierStub{SystemCheckGrantErr: errors.New("denied")}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if RequiredGrant("blogs", "entry", "post", 0)(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}
}

func TestRequireGrantAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/news/1/edit", nil)
	q := &db.QuerierStub{
		SystemCheckGrantFn: func(arg db.SystemCheckGrantParams) (int32, error) {
			if arg.Section != "news" || arg.Action != "edit" {
				t.Fatalf("unexpected section/action: %+v", arg)
			}
			if arg.Item.String != "post" || !arg.Item.Valid {
				t.Fatalf("unexpected item: %+v", arg.Item)
			}
			if arg.ItemID.Int32 != 1 || !arg.ItemID.Valid {
				t.Fatalf("unexpected item id: %+v", arg.ItemID)
			}
			if arg.UserID.Int32 != 1 || !arg.UserID.Valid {
				t.Fatalf("unexpected user id: %+v", arg.UserID)
			}
			if arg.ViewerID != 1 {
				t.Fatalf("unexpected viewer id: %d", arg.ViewerID)
			}
			return 1, nil
		},
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "1"}}
	if !RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to allow request")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check call, got %d", len(q.SystemCheckGrantCalls))
	}
}

func TestRequireGrantDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/news/2/edit", nil)
	q := &db.QuerierStub{
		SystemCheckGrantFn: func(arg db.SystemCheckGrantParams) (int32, error) {
			if arg.ItemID.Int32 != 2 {
				t.Fatalf("unexpected item id: %+v", arg.ItemID)
			}
			return 0, errors.New("denied")
		},
	}
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"user"}))
	cd.UserID = 2
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "2"}}
	if RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to reject request")
	}
	if len(q.SystemCheckGrantCalls) != 1 {
		t.Fatalf("expected 1 grant check call, got %d", len(q.SystemCheckGrantCalls))
	}
}
