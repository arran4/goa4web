package router

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
)

func TestInitModulesOnce(t *testing.T) {
	reg := NewRegistry()
	r := mux.NewRouter()
	count := 0
	reg.RegisterModule("a", nil, func(*mux.Router) { count++ })

	reg.InitModules(r)
	reg.InitModules(r)

	if count != 1 {
		t.Fatalf("expected setup to run once, got %d", count)
	}
}

func TestInitModulesDependencyOrder(t *testing.T) {
	reg := NewRegistry()
	r := mux.NewRouter()
	order := []string{}

	reg.RegisterModule("a", nil, func(*mux.Router) { order = append(order, "a") })
	reg.RegisterModule("b", []string{"a"}, func(*mux.Router) { order = append(order, "b") })
	reg.RegisterModule("c", []string{"b"}, func(*mux.Router) { order = append(order, "c") })

	reg.InitModules(r)

	want := []string{"a", "b", "c"}
	if diff := cmp.Diff(want, order); diff != "" {
		t.Fatalf("order mismatch (-want +got):\n%s", diff)
	}
}
