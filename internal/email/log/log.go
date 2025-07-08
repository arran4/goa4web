package log

import (
	"context"
	"log"
	"net/mail"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider just logs emails for development purposes.
type Provider struct{}

func (Provider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	log.Printf("sending mail to %s\n%s", to.String(), rawEmailMessage)
	return nil
}

func providerFromConfig(runtimeconfig.RuntimeConfig) email.Provider { return Provider{} }

// Register registers the log provider factory.
func Register() { email.RegisterProvider("log", providerFromConfig) }
