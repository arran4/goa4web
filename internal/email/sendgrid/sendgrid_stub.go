//go:build !sendgrid
// +build !sendgrid

package sendgrid

import (
	"context"
	"errors"
	"net/mail"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Built indicates whether the SendGrid provider is compiled in.
const Built = false

// Register registers a stub for the SendGrid provider.
func Register(r *email.Registry) {
	r.RegisterProvider("sendgrid", func(cfg *config.RuntimeConfig) (email.Provider, error) {
		return &stub{}, nil
	})
}

type stub struct{}

func (s *stub) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	return errors.New("sendgrid provider not compiled")
}

func (s *stub) TestConfig(ctx context.Context) error {
	return errors.New("sendgrid provider not compiled")
}
