package news

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathEditRouteRegistered(t *testing.T) {
	r := mux.NewRouter()
	navReg := navpkg.NewRegistry()
	RegisterRoutes(r, config.NewRuntimeConfig(), navReg)

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
