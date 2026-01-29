package router

import (
	"log"
	"sort"
	"sync"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	nav "github.com/arran4/goa4web/internal/navigation"
)

// Module represents a router module and its setup function.
type Module struct {
	Name  string
	Deps  []string
	Setup func(*mux.Router, *config.RuntimeConfig, *nav.Registry)
	once  sync.Once
}

// Registry stores router modules and synchronises access to them.
type Registry struct {
	modules map[string]*Module
	mu      sync.Mutex
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry { return &Registry{modules: map[string]*Module{}} }

// RegisterModule registers a router module with optional dependencies. A module
// is stored only on the first call.
func (reg *Registry) RegisterModule(name string, deps []string, setup func(*mux.Router, *config.RuntimeConfig, *nav.Registry)) {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	if _, ok := reg.modules[name]; ok {
		return
	}
	reg.modules[name] = &Module{Name: name, Deps: deps, Setup: setup}
}

// InitModules initialises all registered modules by resolving their
// dependencies and invoking their Setup function once.
func (reg *Registry) InitModules(r *mux.Router, cfg *config.RuntimeConfig, navReg *nav.Registry) {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	visited := make(map[string]bool)
	var order []*Module

	var visit func(string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		m := reg.modules[name]
		if m == nil {
			log.Printf("Router Registry: Module '%s' not found.", name)
			return
		}
		for _, dep := range m.Deps {
			log.Printf("Router Registry: Module '%s' depends on '%s', visiting...", name, dep)
			visit(dep)
		}
		order = append(order, m)
		log.Printf("Router Registry: Module '%s' resolved.", name)
	}

	for name := range reg.modules {
		visit(name)
	}

	for _, m := range order {
		if m.Setup == nil {
			continue
		}
		m.once.Do(func() {
			log.Printf("Initializing router module: %s", m.Name)
			m.Setup(r, cfg, navReg)
			log.Printf("Initialized router module: %s", m.Name)
		})
	}
}

// Names returns the names of registered modules in sorted order.
func (reg *Registry) Names() []string {
	reg.mu.Lock()
	names := make([]string, 0, len(reg.modules))
	for n := range reg.modules {
		names = append(names, n)
	}
	reg.mu.Unlock()
	sort.Strings(names)
	return names
}
