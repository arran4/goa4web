//go:build !ses
// +build !ses

package ses

import (
	"context"
	"errors"
	"net/mail"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
)

// Built indicates whether the SES provider is compiled in.
const Built = false

// Register registers a stub for the SES provider.
func Register(r *email.Registry) {
	r.RegisterProvider("ses", func(cfg *config.RuntimeConfig) (email.Provider, error) {
		return &stub{}, nil
	})
}

type stub struct{}

func (s *stub) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	return errors.New("ses provider not compiled")
}

func (s *stub) TestConfig(ctx context.Context) error {
	return errors.New("ses provider not compiled")
}
