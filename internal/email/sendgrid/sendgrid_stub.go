//go:build !sendgrid
// +build !sendgrid

package sendgrid

import (
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Built indicates whether the SendGrid provider is compiled in.
const Built = false

// Register registers a stub for the SendGrid provider.
func Register(r *email.Registry) {
	r.RegisterProvider("sendgrid", func(cfg *config.RuntimeConfig) email.Provider {
		return nil
	})
}
