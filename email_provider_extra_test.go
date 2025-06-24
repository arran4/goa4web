package goa4web

import (
	"testing"

	"github.com/arran4/goa4web/internal/email"
)

func TestGetEmailProviderSMTP(t *testing.T) {
	p := providerFromConfig(RuntimeConfig{
		EmailProvider: "smtp",
		EmailSMTPHost: "localhost",
		EmailSMTPPort: "25",
	})
	s, ok := p.(email.SMTPProvider)
	if !ok {
		t.Fatalf("expected SMTPProvider, got %#v", p)
	}
	if s.Addr != "localhost:25" || s.From != email.SourceEmail {
		t.Errorf("unexpected provider values: %#v", s)
	}
}

func TestGetEmailProviderLocal(t *testing.T) {
	if _, ok := providerFromConfig(RuntimeConfig{EmailProvider: "local"}).(email.LocalProvider); !ok {
		t.Fatalf("expected LocalProvider")
	}
}

func TestGetEmailProviderJMAP(t *testing.T) {
	p := providerFromConfig(RuntimeConfig{
		EmailProvider:     "jmap",
		EmailJMAPEndpoint: "http://example.com",
		EmailJMAPAccount:  "acct",
		EmailJMAPIdentity: "id",
	})
	j, ok := p.(email.JMAPProvider)
	if !ok {
		t.Fatalf("expected JMAPProvider, got %#v", p)
	}
	if j.Endpoint != "http://example.com" || j.AccountID != "acct" || j.Identity != "id" {
		t.Errorf("unexpected provider values: %#v", j)
	}
}

func TestGetEmailProviderSESNoCreds(t *testing.T) {
	if p := providerFromConfig(RuntimeConfig{EmailProvider: "ses", EmailAWSRegion: "us-east-1"}); p != nil {
		t.Errorf("expected nil provider, got %#v", p)
	}
}
