package email

import (
	"github.com/arran4/goa4web/config"
	internalemail "github.com/arran4/goa4web/internal/email"
)

// Provider exposes the email.Provider interface for consumers outside the module.
type Provider = internalemail.Provider

// ProviderFromConfig returns an email provider configured from cfg.
func ProviderFromConfig(cfg config.RuntimeConfig) Provider {
	return internalemail.ProviderFromConfig(cfg)
}
