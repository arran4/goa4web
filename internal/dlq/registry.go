package dlq

import (
	"log"
	"strings"
	"sync"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

// ProviderFactory creates a DLQ provider using runtime configuration.
type ProviderFactory func(config.RuntimeConfig, *dbpkg.Queries) DLQ

// Registry holds registered DLQ providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ProviderFactory
}

// NewRegistry returns an initialised DLQ provider registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]ProviderFactory)}
}

// RegisterProvider adds factory to the provider registry under name.
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
