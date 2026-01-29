package router

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/navigation"
)

func TestInitModulesOnce(t *testing.T) {
	reg := NewRegistry()
	r := mux.NewRouter()
	count := 0
	reg.RegisterModule("a", nil, func(*mux.Router, *config.RuntimeConfig, *navigation.Registry) { count++ })

	cfg := &config.RuntimeConfig{}
	navReg := navigation.NewRegistry()
	reg.InitModules(r, cfg, navReg)
	reg.InitModules(r, cfg, navReg)

	if count != 1 {
		t.Fatalf("expected setup to run once, got %d", count)
	}
}

func TestInitModulesDependencyOrder(t *testing.T) {
	reg := NewRegistry()
	r := mux.NewRouter()
	order := []string{}

	navReg := navigation.NewRegistry()

	reg.RegisterModule("a", nil, func(r *mux.Router, cfg *config.RuntimeConfig, _ *navigation.Registry) {
		order = append(order, "a")
	})
	reg.RegisterModule("b", []string{"a"}, func(r *mux.Router, cfg *config.RuntimeConfig, _ *navigation.Registry) {
		order = append(order, "b")
	})
	reg.RegisterModule("c", []string{"b"}, func(r *mux.Router, cfg *config.RuntimeConfig, _ *navigation.Registry) {
		order = append(order, "c")
	})

	reg.InitModules(r, &config.RuntimeConfig{}, navReg)

	want := []string{"a", "b", "c"}
	if diff := cmp.Diff(want, order); diff != "" {
		t.Fatalf("order mismatch (-want +got):\n%s", diff)
	}
}

func TestRegistryNames(t *testing.T) {
	reg := NewRegistry()
	reg.RegisterModule("b", nil, nil)
	reg.RegisterModule("a", nil, nil)
	want := []string{"a", "b"}
	if diff := cmp.Diff(want, reg.Names()); diff != "" {
		t.Fatalf("names mismatch (-want +got):\n%s", diff)
	}
}
