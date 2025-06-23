//go:build !sendgrid
// +build !sendgrid

package main

func sendGridProviderFromConfig(cfg RuntimeConfig) MailProvider {
	return nil
}
