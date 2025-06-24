//go:build !sendgrid
// +build !sendgrid

package goa4web

import "testing"

func TestSendGridProviderUnavailable(t *testing.T) {
	if p := providerFromConfig(RuntimeConfig{EmailProvider: "sendgrid", EmailSendGridKey: "k"}); p != nil {
		t.Fatalf("expected nil provider when sendgrid tag not enabled")
	}
}
