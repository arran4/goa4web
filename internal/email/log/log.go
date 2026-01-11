package log

import (
	"context"
	"log"
	"net/mail"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Provider just logs emails for development purposes.
type Provider struct{ Verbosity int }

func (p Provider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	if p.Verbosity >= email.LogLevelBody {
		log.Printf("sending mail to %s\n%s", to.String(), rawEmailMessage)
	} else if p.Verbosity >= email.LogLevelSummary {
		log.Printf("sending mail to %s", to.String())
	}
	return nil
}

func (p Provider) TestConfig(ctx context.Context) error {
	log.Printf("Log provider is enabled with verbosity %d", p.Verbosity)
	return nil
}

func providerFromConfig(cfg *config.RuntimeConfig) (email.Provider, error) {
	return Provider{Verbosity: cfg.EmailLogVerbosity}, nil
}

// Register registers the log provider factory.
func Register(r *email.Registry) { r.RegisterProvider("log", providerFromConfig) }
