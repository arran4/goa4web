package email

import (
	internalemail "github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider exposes the email.Provider interface for consumers outside the module.
type Provider = internalemail.Provider

// ProviderFromConfig returns an email provider configured from cfg.
func ProviderFromConfig(cfg runtimeconfig.RuntimeConfig) Provider {
	return internalemail.ProviderFromConfig(cfg)
}
