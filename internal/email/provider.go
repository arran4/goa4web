package email

import (
	"context"
	"net/mail"
)

// Provider defines a simple interface that all mail backends must implement.
// Only the fields necessary for sending basic notification emails are included.
type Provider interface {
	Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error
	TestConfig(ctx context.Context) (string, error)
}
