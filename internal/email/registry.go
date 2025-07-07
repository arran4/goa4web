package email

import (
	"log"
	"strings"

	"github.com/arran4/goa4web/runtimeconfig"
)

var providerRegistry = map[string]func(runtimeconfig.RuntimeConfig) Provider{}

// RegisterProvider registers a factory for name.
func RegisterProvider(name string, factory func(runtimeconfig.RuntimeConfig) Provider) {
	n := strings.ToLower(name)
	if _, ok := providerRegistry[n]; ok {
		log.Printf("email: provider %s already registered", n)
	}
	providerRegistry[n] = factory
}

// providerFactory looks up the factory for name.
func providerFactory(name string) func(runtimeconfig.RuntimeConfig) Provider {
	return providerRegistry[strings.ToLower(name)]
}
