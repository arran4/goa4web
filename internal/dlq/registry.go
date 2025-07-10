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
