package news

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
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

func TestCustomNewsIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	CustomNewsIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin not in admin mode should not see add news")
	}

	cd.AdminMode = true
	CustomNewsIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("admin should see add news when admin mode is enabled")
	}

	ctx := req.Context()
	cd = common.NewCoreData(ctx, grantQueries{allowed: true}, config.NewRuntimeConfig(), common.WithUserRoles([]string{"content writer"}))
	cd.UserID = 1
	CustomNewsIndex(cd, req.WithContext(ctx))
	if !common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("content writer with grant should see add news")
	}

	cd = common.NewCoreData(req.Context(), grantQueries{}, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	CustomNewsIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Add News") {
		t.Errorf("user without grant should not see add news")
	}
}
