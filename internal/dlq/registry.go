package dlq

import (
	"sync"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
)

// ProviderFactory creates a DLQ provider using runtime configuration.
type ProviderFactory func(runtimeconfig.RuntimeConfig, *dbpkg.Queries) DLQ

var (
	regMu     sync.RWMutex
	providers = make(map[string]ProviderFactory)
)

// RegisterProvider adds factory to the provider registry under name.
func RegisterProvider(name string, factory ProviderFactory) {
	regMu.Lock()
	defer regMu.Unlock()
	providers[name] = factory
}

// lookupProvider retrieves a factory by name.
func lookupProvider(name string) ProviderFactory {
	regMu.RLock()
	f := providers[name]
	regMu.RUnlock()
	return f
}
