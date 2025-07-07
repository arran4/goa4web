package log

import (
	"context"
	"log"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider just logs emails for development purposes.
type Provider struct{}

func (Provider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	if htmlBody != "" {
		log.Printf("sending mail to %s subject %q\nTEXT:\n%s\nHTML:\n%s", to, subject, textBody, htmlBody)
	} else {
		log.Printf("sending mail to %s subject %q\n%s", to, subject, textBody)
	}
	return nil
}

func providerFromConfig(runtimeconfig.RuntimeConfig) email.Provider { return Provider{} }

// Register registers the log provider factory.
func Register() { email.RegisterProvider("log", providerFromConfig) }
