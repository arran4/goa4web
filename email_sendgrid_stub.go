//go:build !sendgrid
// +build !sendgrid

package goa4web

func sendGridProviderFromConfig(cfg RuntimeConfig) MailProvider {
	return nil
}
