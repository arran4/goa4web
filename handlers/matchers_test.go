package handlers

import (
	"context"
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
