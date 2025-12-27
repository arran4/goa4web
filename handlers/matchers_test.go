package handlers

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type grantQueries struct {
	db.Querier
	allowed bool
}

func (g grantQueries) SystemCheckGrant(context.Context, db.SystemCheckGrantParams) (int32, error) {
	if g.allowed {
		return 1, nil
	}
	return 0, sql.ErrNoRows
}

func (g grantQueries) SystemCheckRoleGrant(context.Context, db.SystemCheckRoleGrantParams) (int32, error) {
	return 0, sql.ErrNoRows
}

func TestRequiredAccessAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"content writer"}))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if !RequiredAccess("content writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access allowed")
	}
}

func TestRequiredAccessDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if RequiredAccess("content writer")(req, &mux.RouteMatch{}) {
		t.Errorf("expected access denied")
	}
}

func TestRequireGrantAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/news/1/edit", nil)
	cd := common.NewCoreData(req.Context(), grantQueries{allowed: true}, config.NewRuntimeConfig(), common.WithUserRoles([]string{}))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "1"}}
	if !RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to allow request")
	}
}

func TestRequireGrantDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/news/2/edit", nil)
	cd := common.NewCoreData(req.Context(), grantQueries{}, config.NewRuntimeConfig(), common.WithUserRoles([]string{}))
	cd.UserID = 2
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	match := &mux.RouteMatch{Vars: map[string]string{"news": "2"}}
	if RequireGrantForPathInt("news", "post", "edit", "news")(req, match) {
		t.Fatalf("expected grant-based matcher to reject request")
	}
}
