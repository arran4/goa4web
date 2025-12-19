package email

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/arran4/goa4web/config"
)

// ProviderFactory creates a mail provider from cfg.
type ProviderFactory func(*config.RuntimeConfig) (Provider, error)

// Registry stores email provider factories.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ProviderFactory
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{providers: make(map[string]ProviderFactory)} }

// RegisterProvider registers factory under name.
func (r *Registry) RegisterProvider(name string, factory ProviderFactory) {
	n := strings.ToLower(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[n]; ok {
		log.Printf("email: provider %s already registered", n)
	}
	r.providers[n] = factory
}

// providerFactory retrieves the factory for name.
func (r *Registry) providerFactory(name string) ProviderFactory {
	r.mu.RLock()
	f := r.providers[strings.ToLower(name)]
	r.mu.RUnlock()
	return f
}

// ProviderFromConfig returns a provider configured from cfg.
func (r *Registry) ProviderFromConfig(cfg *config.RuntimeConfig) (Provider, error) {
	mode := strings.ToLower(cfg.EmailProvider)
	if f := r.providerFactory(mode); f != nil {
		return f(cfg)
	}
	if mode != "" {
		return nil, fmt.Errorf("Email disabled: unknown provider %q", mode)
	}
	return nil, nil
}

// ProviderNames returns registered provider names in sorted order.
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
