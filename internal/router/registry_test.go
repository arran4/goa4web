package router

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
)

func TestInitModulesOnce(t *testing.T) {
	reg := NewRegistry()
	r := mux.NewRouter()
	count := 0
	reg.RegisterModule("a", nil, func(*mux.Router, *config.RuntimeConfig) { count++ })

	cfg := &config.RuntimeConfig{}
	reg.InitModules(r, cfg)
	reg.InitModules(r, cfg)

	if count != 1 {
		t.Fatalf("expected setup to run once, got %d", count)
	}
}

func TestInitModulesDependencyOrder(t *testing.T) {
	reg := NewRegistry()
	r := mux.NewRouter()
	order := []string{}

	reg.RegisterModule("a", nil, func(*mux.Router, *config.RuntimeConfig) { order = append(order, "a") })
	reg.RegisterModule("b", []string{"a"}, func(*mux.Router, *config.RuntimeConfig) { order = append(order, "b") })
	reg.RegisterModule("c", []string{"b"}, func(*mux.Router, *config.RuntimeConfig) { order = append(order, "c") })

	reg.InitModules(r, &config.RuntimeConfig{})

	want := []string{"a", "b", "c"}
	if diff := cmp.Diff(want, order); diff != "" {
		t.Fatalf("order mismatch (-want +got):\n%s", diff)
	}
}
