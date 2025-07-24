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

var (
	regMu     sync.RWMutex
	providers = make(map[string]ProviderFactory)
)

// RegisterProvider adds factory to the provider registry under name.
func RegisterProvider(name string, factory ProviderFactory) {
	regMu.Lock()
	defer regMu.Unlock()
	n := strings.ToLower(name)
	if _, ok := providers[n]; ok {
		log.Printf("dlq: provider %s already registered", n)
	}
	providers[n] = factory
}

// lookupProvider retrieves a factory by name.
func lookupProvider(name string) ProviderFactory {
	regMu.RLock()
	f := providers[strings.ToLower(name)]
	regMu.RUnlock()
	return f
}

// ProviderNames returns the names of registered DLQ providers in sorted order.
func ProviderNames() []string {
	regMu.RLock()
	names := make([]string, 0, len(providers))
	for n := range providers {
		names = append(names, n)
	}
	regMu.RUnlock()
	sort.Strings(names)
	return names
}
