package main

import "testing"

func TestGetEmailProviderSMTP(t *testing.T) {
	p := providerFromConfig(EmailConfig{
		Provider: "smtp",
		SMTPHost: "localhost",
		SMTPPort: "25",
	})
	s, ok := p.(smtpMailProvider)
	if !ok {
		t.Fatalf("expected smtpMailProvider, got %#v", p)
	}
	if s.addr != "localhost:25" || s.from != SourceEmail {
		t.Errorf("unexpected provider values: %#v", s)
	}
}

func TestGetEmailProviderLocal(t *testing.T) {
	if _, ok := providerFromConfig(EmailConfig{Provider: "local"}).(localMailProvider); !ok {
		t.Fatalf("expected localMailProvider")
	}
}

func TestGetEmailProviderJMAP(t *testing.T) {
	p := providerFromConfig(EmailConfig{
		Provider:     "jmap",
		JMAPEndpoint: "http://example.com",
		JMAPAccount:  "acct",
		JMAPIdentity: "id",
	})
	j, ok := p.(jmapMailProvider)
	if !ok {
		t.Fatalf("expected jmapMailProvider, got %#v", p)
	}
	if j.endpoint != "http://example.com" || j.accountID != "acct" || j.identity != "id" {
		t.Errorf("unexpected provider values: %#v", j)
	}
}

func TestGetEmailProviderSESNoCreds(t *testing.T) {
	if p := providerFromConfig(EmailConfig{Provider: "ses", AWSRegion: "us-east-1"}); p != nil {
		t.Errorf("expected nil provider, got %#v", p)
	}
}
