//go:build !sendgrid
// +build !sendgrid

package goa4web

import "context"

// sendgridBuilt indicates whether the SendGrid provider is compiled in.
const sendgridBuilt = false

// sendGridProvider is a stub implementation used when SendGrid support is disabled.
type sendGridProvider struct{}

func (sendGridProvider) Send(ctx context.Context, to, subject, body string) error {
	return nil
}

func sendGridProviderFromConfig(cfg RuntimeConfig) MailProvider {
	return nil
}
