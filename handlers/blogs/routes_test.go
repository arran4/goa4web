package blogs

import (
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	navpkg "github.com/arran4/goa4web/internal/navigation"
)

func TestRegisterRoutesRegistersAdminLink(t *testing.T) {
	r := mux.NewRouter()
	navReg := navpkg.NewRegistry()
	RegisterRoutes(r, config.NewRuntimeConfig(), navReg, nil, nil)
	links := navReg.AdminLinks()
	for _, l := range links {
		if l.Name == "Blogs" {
			if l.Link != "/admin/blogs" {
				t.Fatalf("expected /admin/blogs got %s", l.Link)
			}
			return
		}
	}
	t.Fatalf("Blogs link not registered")
}
