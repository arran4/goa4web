package email

import (
	"log"
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
