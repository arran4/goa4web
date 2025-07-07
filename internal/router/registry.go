package router

import (
	"sync"

	"github.com/gorilla/mux"
)

// Module represents a router module and its setup function.
type Module struct {
	Name  string
	Deps  []string
	Setup func(*mux.Router)
	once  sync.Once
}

var (
	modules = map[string]*Module{}
	mu      sync.Mutex
)

// RegisterModule registers a router module with optional dependencies. A module
// is stored only on the first call.
func RegisterModule(name string, deps []string, setup func(*mux.Router)) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := modules[name]; ok {
		return
	}
	modules[name] = &Module{Name: name, Deps: deps, Setup: setup}
}

// InitModules initialises all registered modules by resolving their
// dependencies and invoking their Setup function once.
func InitModules(r *mux.Router) {
	mu.Lock()
	defer mu.Unlock()

	visited := make(map[string]bool)
	var order []*Module

	var visit func(string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		m := modules[name]
		if m == nil {
			return
		}
		for _, dep := range m.Deps {
			visit(dep)
		}
		order = append(order, m)
	}

	for name := range modules {
		visit(name)
	}

	for _, m := range order {
		if m.Setup == nil {
			continue
		}
		m.once.Do(func() { m.Setup(r) })
	}
}
