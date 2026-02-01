package email

import (
	"context"
	"fmt"
	"log"
	"net/mail"
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
	var p Provider
	var err error

	if f := r.providerFactory(mode); f != nil {
		p, err = f(cfg)
	} else if mode != "" {
		return nil, fmt.Errorf("Email disabled: unknown provider %q", mode)
	}

	if err != nil {
		return nil, err
	}

	if p != nil && cfg.EmailOverride != "" {
		p = &overrideProvider{
			Provider: p,
			override: cfg.EmailOverride,
		}
	}
	return p, nil
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

type overrideProvider struct {
	Provider
	override string
}

func (p *overrideProvider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	log.Printf("email: overriding recipient %s with %s", to.Address, p.override)
	to = mail.Address{Address: p.override}
	return p.Provider.Send(ctx, to, rawEmailMessage)
}
