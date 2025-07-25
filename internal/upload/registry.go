package upload

import (
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/arran4/goa4web/config"
)

// ProviderFactory creates an upload provider from cfg.
type ProviderFactory func(*config.RuntimeConfig) Provider

var (
	regMu    sync.RWMutex
	registry = map[string]ProviderFactory{}
)

// RegisterProvider adds factory to the provider registry under name.
func RegisterProvider(name string, factory ProviderFactory) {
	regMu.Lock()
	defer regMu.Unlock()
	n := strings.ToLower(name)
	if _, ok := registry[n]; ok {
		log.Printf("upload: provider %s already registered", n)
	}
	registry[n] = factory
}

func providerFactory(name string) ProviderFactory {
	regMu.RLock()
	f := registry[strings.ToLower(name)]
	regMu.RUnlock()
	return f
}

// ProviderNames returns the names of registered upload providers in sorted order.
func ProviderNames() []string {
	regMu.RLock()
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	regMu.RUnlock()
	sort.Strings(names)
	return names
}
