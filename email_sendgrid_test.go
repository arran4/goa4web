//go:build sendgrid
// +build sendgrid

package goa4web

import "testing"

func TestSendGridProviderFromConfig(t *testing.T) {
	p := providerFromConfig(RuntimeConfig{EmailProvider: "sendgrid", EmailSendGridKey: "k"})
	if _, ok := p.(sendGridProvider); !ok {
		t.Fatalf("expected sendGridProvider, got %#v", p)
	}
}
