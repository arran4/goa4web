package news

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathEditRouteRegistered(t *testing.T) {
	r := mux.NewRouter()
	RegisterRoutes(r, config.NewRuntimeConfig())

	found := false
	_ = r.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		if path == "/news/news/{news}/edit" {
			methods := testhelpers.Must(route.GetMethods())
			for _, m := range methods {
				if m == http.MethodGet {
					found = true
				}
			}
		}
		return nil
	})
	if !found {
		t.Fatalf("expected edit route to be registered")
	}
}

func TestNewsDetailRouteChecksViewGrantForRequestedID(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.SystemCheckGrantFn = func(params db.SystemCheckGrantParams) (int32, error) {
		return 0, sql.ErrNoRows
	}
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 7

	r := mux.NewRouter()
	RegisterRoutes(r, config.NewRuntimeConfig())
	req := httptest.NewRequest(http.MethodGet, "/news/news/42", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	var match mux.RouteMatch
	if !r.Match(req, &match) {
		t.Fatal("news detail route did not match")
	}
	if got := match.Vars["news"]; got != "42" {
		t.Fatalf("matched news ID=%q, want 42", got)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusForbidden)
	}
	if len(queries.SystemCheckGrantCalls) == 0 {
		t.Fatal("news detail route did not check a view grant")
	}
	foundViewCheck := false
	for _, grant := range queries.SystemCheckGrantCalls {
		if grant.Section == "news" && grant.Item.Valid && grant.Item.String == "post" && grant.Action == "view" {
			foundViewCheck = true
			if !grant.ItemID.Valid || grant.ItemID.Int32 != 42 {
				t.Fatalf("unexpected direct-detail view grant check: %#v", grant)
			}
		}
	}
	if !foundViewCheck {
		t.Fatalf("news detail route did not check view for requested item: %#v", queries.SystemCheckGrantCalls)
	}
}
