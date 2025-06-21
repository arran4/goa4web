//go:build sendgrid
// +build sendgrid

package main

import "testing"

func TestSendGridProviderFromConfig(t *testing.T) {
	p := providerFromConfig(EmailConfig{Provider: "sendgrid", SendGridKey: "k"})
	if _, ok := p.(sendGridProvider); !ok {
		t.Fatalf("expected sendGridProvider, got %#v", p)
	}
}
