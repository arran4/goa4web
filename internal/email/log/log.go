package log

import (
	"context"
	"log"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider just logs emails for development purposes.
type Provider struct{}

func (Provider) Send(ctx context.Context, to, subject string, rawEmailMessage []byte) error {
	log.Printf("sending mail to %s subject %q\n%s", to, subject, rawEmailMessage)
	return nil
}

func providerFromConfig(runtimeconfig.RuntimeConfig) email.Provider { return Provider{} }

// Register registers the log provider factory.
func Register() { email.RegisterProvider("log", providerFromConfig) }
