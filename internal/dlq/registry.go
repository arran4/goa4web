package dlq

import (
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

// ProviderFactory creates a DLQ provider using runtime configuration.
type ProviderFactory func(config.RuntimeConfig, *dbpkg.Queries) DLQ

// Registry stores DLQ provider factories.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ProviderFactory
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{providers: make(map[string]ProviderFactory)} }

// RegisterProvider adds factory to the registry under name.
func (r *Registry) RegisterProvider(name string, factory ProviderFactory) {
	n := strings.ToLower(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[n]; ok {
		log.Printf("dlq: provider %s already registered", n)
	}
	r.providers[n] = factory
}

// lookupProvider retrieves a factory by name.
func (r *Registry) lookupProvider(name string) ProviderFactory {
	r.mu.RLock()
	f := r.providers[strings.ToLower(name)]
	r.mu.RUnlock()
	return f
}

// ProviderFromConfig returns a DLQ implementation configured from cfg.
func (r *Registry) ProviderFromConfig(cfg config.RuntimeConfig, q *dbpkg.Queries) DLQ {
	names := strings.Split(cfg.DLQProvider, ",")
	var qs []DLQ
	for _, name := range names {
		n := strings.ToLower(strings.TrimSpace(name))
		if n == "" {
			continue
		}
		if f := r.lookupProvider(n); f != nil {
			qs = append(qs, f(cfg, q))
		} else {
			if n != "log" {
				log.Printf("unrecognised DLQ provider %q, falling back to log", n)
			}
			qs = append(qs, LogDLQ{})
		}
	}
	if len(qs) == 0 {
		return LogDLQ{}
	}
	if len(qs) == 1 {
		return qs[0]
	}
	return NewMulti(qs...)
}

// ProviderNames returns the names of registered DLQ providers in sorted order.
func (r *Registry) ProviderNames() []string {
	r.mu.RLock()
	names := make([]string, 0, len(r.providers))
	for n := range r.providers {
		names = append(names, n)
	}
	r.mu.RUnlock()
	sort.Strings(names)
	return names
}

// DefaultRegistry holds the default providers.
var DefaultRegistry = NewRegistry()

// RegisterProvider registers factory in the default registry.
func RegisterProvider(name string, factory ProviderFactory) {
	DefaultRegistry.RegisterProvider(name, factory)
}

// ProviderFromConfig returns a provider from the default registry.
func ProviderFromConfig(cfg config.RuntimeConfig, q *dbpkg.Queries) DLQ {
	return DefaultRegistry.ProviderFromConfig(cfg, q)
}

// ProviderNames lists providers in the default registry.
func ProviderNames() []string { return DefaultRegistry.ProviderNames() }
