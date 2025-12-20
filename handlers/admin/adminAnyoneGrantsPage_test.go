package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type anyoneGrantsQueries struct {
	db.Querier
	grants []*db.Grant
}

func (q *anyoneGrantsQueries) ListGrants(context.Context) ([]*db.Grant, error) {
	return q.grants, nil
}

func TestAdminAnyoneGrantsPage(t *testing.T) {
	queries := &anyoneGrantsQueries{
		grants: []*db.Grant{{
			ID:       1,
			Section:  "forum",
			Item:     sql.NullString{},
			RuleType: "allow",
			Action:   "search",
			Active:   true,
		}},
	}

	req := httptest.NewRequest("GET", "/admin/grants/anyone", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	AdminAnyoneGrantsPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `<a href="/admin/grants/anyone">Anyone</a>`) {
		t.Fatalf("missing link: %s", body)
	}
	if !strings.Contains(body, `<a href="/admin/grant/1" class="pill">search</a>`) {
		t.Fatalf("missing action pill: %s", body)
	}
}
