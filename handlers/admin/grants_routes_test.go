package admin

import (
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestHappyPathRegisterRoutesRegistersGrantsLink(t *testing.T) {
	h := New()
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	navReg := navpkg.NewRegistry()
	h.RegisterRoutes(ar, &config.RuntimeConfig{}, navReg)
	links := navReg.AdminLinks()
	for _, l := range links {
		if l.Name == "Grants" {
			if l.Link != "/admin/grants" {
				t.Fatalf("expected /admin/grants got %s", l.Link)
			}
			return
		}
	}
	t.Fatalf("Grants link not registered")
}

func TestHappyPathRegisterRoutesRegistersGrantAdd(t *testing.T) {
	h := New()
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	navReg := navpkg.NewRegistry()
	h.RegisterRoutes(ar, &config.RuntimeConfig{}, navReg)
	var found bool
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		if path == "/admin/grant/add" {
			methods := testhelpers.Must(route.GetMethods())
			for _, m := range methods {
				if m == "GET" {
					found = true
				}
			}
		}
		return nil
	})
	if !found {
		t.Fatalf("grant add route not registered")
	}
}

func TestHappyPathRegisterRoutesRegistersGrantCreate(t *testing.T) {
	h := New()
	r := mux.NewRouter()
	ar := r.PathPrefix("/admin").Subrouter()
	navReg := navpkg.NewRegistry()
	h.RegisterRoutes(ar, &config.RuntimeConfig{}, navReg)
	var found bool
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		if path == "/admin/grant" {
			methods := testhelpers.Must(route.GetMethods())
			for _, m := range methods {
				if m == "POST" {
					found = true
				}
			}
		}
		return nil
	})
	if !found {
		t.Fatalf("grant create route not registered")
	}
}
