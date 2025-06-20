//go:build !sendgrid
// +build !sendgrid

package main

func sendGridProviderFromConfig(cfg EmailConfig) MailProvider {
	return nil
}
