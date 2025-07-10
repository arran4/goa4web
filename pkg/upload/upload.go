package upload

import (
	internal "github.com/arran4/goa4web/internal/upload"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider exposes the upload.Provider interface externally.
type Provider = internal.Provider

// CacheProvider exposes the upload.CacheProvider interface externally.
type CacheProvider = internal.CacheProvider

// ProviderFromConfig returns a provider selected by runtime configuration.
func ProviderFromConfig(cfg runtimeconfig.RuntimeConfig) Provider {
	return internal.ProviderFromConfig(cfg)
}

// CacheProviderFromConfig returns the cache provider selected by runtime configuration.
func CacheProviderFromConfig(cfg runtimeconfig.RuntimeConfig) Provider {
	return internal.CacheProviderFromConfig(cfg)
}
