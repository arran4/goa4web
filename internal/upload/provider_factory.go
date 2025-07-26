package upload

import "strings"

import "github.com/arran4/goa4web/config"

// ProviderFromConfig returns a provider selected by cfg.ImageUploadProvider.
func ProviderFromConfig(cfg *config.RuntimeConfig) Provider {
	name := strings.ToLower(cfg.ImageUploadProvider)
	if f := providerFactory(name); f != nil {
		return f(cfg)
	}
	return nil
}

// CacheProviderFromConfig returns a provider selected by cfg.ImageCacheProvider.
func CacheProviderFromConfig(cfg *config.RuntimeConfig) Provider {
	c := *cfg
	c.ImageUploadProvider = cfg.ImageCacheProvider
	c.ImageUploadDir = cfg.ImageCacheDir
	c.ImageUploadS3URL = cfg.ImageCacheS3URL
	return ProviderFromConfig(&c)
}
