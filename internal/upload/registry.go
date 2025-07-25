package upload

import (
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/arran4/goa4web/config"
)

// ProviderFactory creates an upload provider from cfg.
type ProviderFactory func(config.RuntimeConfig) Provider

// Registry stores upload provider factories.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ProviderFactory
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry { return &Registry{providers: make(map[string]ProviderFactory)} }

// RegisterProvider adds factory to the Registry under name.
func (r *Registry) RegisterProvider(name string, factory ProviderFactory) {
	n := strings.ToLower(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[n]; ok {
		log.Printf("upload: provider %s already registered", n)
	}
	r.providers[n] = factory
}

func (r *Registry) providerFactory(name string) ProviderFactory {
	r.mu.RLock()
	f := r.providers[strings.ToLower(name)]
	r.mu.RUnlock()
	return f
}

// ProviderFromConfig returns a provider selected by cfg.ImageUploadProvider.
func (r *Registry) ProviderFromConfig(cfg config.RuntimeConfig) Provider {
	name := strings.ToLower(cfg.ImageUploadProvider)
	if f := r.providerFactory(name); f != nil {
		return f(cfg)
	}
	return nil
}

// CacheProviderFromConfig returns a provider selected by cfg.ImageCacheProvider.
func (r *Registry) CacheProviderFromConfig(cfg config.RuntimeConfig) Provider {
	c := cfg
	c.ImageUploadProvider = cfg.ImageCacheProvider
	c.ImageUploadDir = cfg.ImageCacheDir
	c.ImageUploadS3URL = cfg.ImageCacheS3URL
	return r.ProviderFromConfig(c)
}

// ProviderNames returns the names of registered upload providers in sorted order.
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
