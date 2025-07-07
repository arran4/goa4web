package router

import (
	"reflect"
	"testing"

	"github.com/gorilla/mux"
)

func TestInitModulesOnce(t *testing.T) {
	modules = map[string]*Module{}
	t.Cleanup(func() { modules = map[string]*Module{} })
	r := mux.NewRouter()
	count := 0
	RegisterModule("a", nil, func(*mux.Router) { count++ })

	InitModules(r)
	InitModules(r)

	if count != 1 {
		t.Fatalf("expected setup to run once, got %d", count)
	}
}

func TestInitModulesDependencyOrder(t *testing.T) {
	modules = map[string]*Module{}
	t.Cleanup(func() { modules = map[string]*Module{} })
	r := mux.NewRouter()
	order := []string{}

	RegisterModule("a", nil, func(*mux.Router) { order = append(order, "a") })
	RegisterModule("b", []string{"a"}, func(*mux.Router) { order = append(order, "b") })
	RegisterModule("c", []string{"b"}, func(*mux.Router) { order = append(order, "c") })

	InitModules(r)

	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order %+v want %+v", order, want)
	}
}
