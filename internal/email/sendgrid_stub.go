//go:build !sendgrid
// +build !sendgrid

package email

import "context"

// SendgridBuilt indicates whether the SendGrid provider is compiled in.
const SendgridBuilt = false

// SendGridProvider is a stub implementation used when SendGrid support is disabled.
type SendGridProvider struct{}

func (SendGridProvider) Send(ctx context.Context, to, subject, body string) error { return nil }

func SendGridProviderFromConfig(key string) Provider { return nil }
