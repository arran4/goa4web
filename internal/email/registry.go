package email

import (
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/arran4/goa4web/config"
)

var (
	regMu            sync.RWMutex
	providerRegistry = map[string]func(config.RuntimeConfig) Provider{}
)

// RegisterProvider registers a factory for name.
func RegisterProvider(name string, factory func(config.RuntimeConfig) Provider) {
	n := strings.ToLower(name)
	regMu.Lock()
	defer regMu.Unlock()
	if _, ok := providerRegistry[n]; ok {
		log.Printf("email: provider %s already registered", n)
	}
	providerRegistry[n] = factory
}

// providerFactory looks up the factory for name.
func providerFactory(name string) func(config.RuntimeConfig) Provider {
	regMu.RLock()
	f := providerRegistry[strings.ToLower(name)]
	regMu.RUnlock()
	return f
}

// ProviderNames returns the names of registered email providers in sorted order.
func ProviderNames() []string {
	regMu.RLock()
	names := make([]string, 0, len(providerRegistry))
	for n := range providerRegistry {
		names = append(names, n)
	}
	regMu.RUnlock()
	sort.Strings(names)
	return names
}
