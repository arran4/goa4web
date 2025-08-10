package forum

import (
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/gorilla/mux"
)

// TestCategoryRoute verifies that the public category route exists.
func TestCategoryRoute(t *testing.T) {
	r := mux.NewRouter()
	RegisterRoutes(r, &config.RuntimeConfig{}, navpkg.NewRegistry())
	req := httptest.NewRequest("GET", "/forum/category/1", nil)
	m := &mux.RouteMatch{}
	if !r.Match(req, m) || m.Handler == nil {
		t.Fatalf("route /forum/category/{id} not registered")
	}
}
