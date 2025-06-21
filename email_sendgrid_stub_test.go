//go:build !sendgrid
// +build !sendgrid

package main

import "testing"

func TestSendGridProviderUnavailable(t *testing.T) {
	if p := providerFromConfig(EmailConfig{Provider: "sendgrid", SendGridKey: "k"}); p != nil {
		t.Fatalf("expected nil provider when sendgrid tag not enabled")
	}
}
