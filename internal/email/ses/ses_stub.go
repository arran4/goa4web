//go:build !ses

package ses

import (
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Built indicates whether the SES provider is compiled in.
const Built = false

func providerFromConfig(*config.RuntimeConfig) email.Provider { return nil }

// Register is a no-op when the ses build tag is not present.
func Register(r *email.Registry) { r.RegisterProvider("ses", providerFromConfig) }
