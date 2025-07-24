package email

import (
	"log"
	"strings"
	"sync"

	"github.com/arran4/goa4web/config"
)

// Registry stores registered email providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]func(config.RuntimeConfig) Provider
}

// NewRegistry returns an initialised email provider registry.
func NewRegistry() *Registry {
	return &Registry{providers: map[string]func(config.RuntimeConfig) Provider{}}
}

// RegisterProvider registers factory under name.
func (r *Registry) RegisterProvider(name string, factory func(config.RuntimeConfig) Provider) {
	n := strings.ToLower(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[n]; ok {
		log.Printf("email: provider %s already registered", n)
	}
	r.providers[n] = factory
}

// providerFactory looks up the factory for name.
func (r *Registry) providerFactory(name string) func(config.RuntimeConfig) Provider {
	r.mu.RLock()
	f := r.providers[strings.ToLower(name)]
	r.mu.RUnlock()
	return f
}
