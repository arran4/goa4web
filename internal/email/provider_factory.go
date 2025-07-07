package email

import (
	"log"
	"strings"

	"github.com/arran4/goa4web/runtimeconfig"
)

// ProviderFromConfig returns an email provider configured from cfg.
func ProviderFromConfig(cfg runtimeconfig.RuntimeConfig) Provider {
	mode := strings.ToLower(cfg.EmailProvider)

	if f := providerFactory(mode); f != nil {
		return f(cfg)
	}

	if mode != "" {
		log.Printf("Email disabled: unknown provider %q", mode)
	}
	return nil
}
